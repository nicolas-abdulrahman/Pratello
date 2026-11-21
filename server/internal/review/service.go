package review

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	errInvalid     = errors.New("invalid review")
	errMissing     = errors.New("review target or purchase not found")
	errForbidden   = errors.New("this purchase does not authorize a review")
	errConflict    = errors.New("purchase already reviewed or not eligible for review")
	errUnavailable = errors.New("review database unavailable")
)

type Review struct {
	ID              int64     `json:"id"`
	Rating          int       `json:"rating"`
	Comment         *string   `json:"comment,omitempty"`
	Created         time.Time `json:"created_at"`
	AuthorName      string    `json:"author_name,omitempty"`
	AuthorAvatarURL string    `json:"author_avatar_url,omitempty"`
}

type totals struct {
	Count   int64    `json:"review_count"`
	Sum     int64    `json:"rating_sum"`
	Average *float64 `json:"average_rating"`
	Sold    *int64   `json:"units_sold,omitempty"`
}

type page struct {
	Data    []Review `json:"data"`
	Summary totals   `json:"summary"`
	Page    int      `json:"page"`
	Size    int      `json:"page_size"`
}

type input struct {
	Rating  int     `json:"rating"`
	Comment *string `json:"comment"`
}

type request struct {
	item                                   bool
	slug, guest                            string
	actor, order, line, target, page, size int
	input                                  input
}

type purchase struct {
	owner, target       int
	guestHash           []byte
	delivered, eligible bool
}

type unit interface {
	publicTarget(context.Context, request) (int, error)
	totals(context.Context, bool, int) (totals, error)
	list(context.Context, bool, int, int, int64) ([]Review, error)
	purchase(context.Context, request) (purchase, error)
	save(context.Context, request, int) (Review, error)
}

type repository interface {
	transact(context.Context, bool, func(unit) (any, error)) (any, error)
}

type Service struct{ store repository }

func (s *Service) list(ctx context.Context, r request) (any, error) {
	return s.store.transact(ctx, false, func(u unit) (any, error) {
		id, err := u.publicTarget(ctx, r)
		if err != nil {
			return nil, err
		}
		summary, err := u.totals(ctx, r.item, id)
		if err != nil {
			return nil, err
		}
		if summary.Count > 0 {
			mean := float64(summary.Sum) / float64(summary.Count)
			summary.Average = &mean
		}
		rows, err := u.list(ctx, r.item, id, r.size, int64(r.page-1)*int64(r.size))
		if err != nil {
			return nil, err
		}
		return page{rows, summary, r.page, r.size}, nil
	})
}

func (s *Service) write(ctx context.Context, r request) (any, error) {
	limit := 5
	if r.item {
		limit = 3
	}
	if r.input.Rating < 1 || r.input.Rating > limit {
		return nil, errInvalid
	}
	if r.input.Comment != nil {
		if r.actor == 0 {
			return nil, errForbidden
		}
		comment := strings.TrimSpace(*r.input.Comment)
		if utf8.RuneCountInString(comment) > 2000 || strings.ContainsRune(comment, 0) {
			return nil, errInvalid
		}
		r.input.Comment = nil
		if comment != "" {
			r.input.Comment = &comment
		}
	}
	return s.store.transact(ctx, true, func(u unit) (any, error) {
		p, err := u.purchase(ctx, r)
		if err != nil {
			return nil, err
		}
		if !ownsPurchase(p, r.actor, r.guest) {
			return nil, errForbidden
		}
		if !p.delivered || !p.eligible {
			return nil, errConflict
		}
		return u.save(ctx, r, p.target)
	})
}

func ownsPurchase(p purchase, actor int, token string) bool {
	if actor > 0 {
		return p.owner == actor
	}
	if p.owner != 0 || len(token) != 43 {
		return false
	}
	hash := sha256.Sum256([]byte(token))
	return subtle.ConstantTimeCompare(p.guestHash, hash[:]) == 1
}
