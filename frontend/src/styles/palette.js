/**
 * Pratello Color Palette & Theme Definitions
 *
 * - restaurant theme navbar: #e57442 (softer, desaturated warm orange)
 * - brand theme navbar: #FAF4EA (soft banana)
 */

export const palette = {
  // Brand: soft banana & warm amber (used for RestaurantsPage)
  brand: {
    primary: '#f59e0b',
    primaryHover: '#d97706',
    primarySoft: '#fef3c7',
    bg: '#FAF6F0',
    surface: '#FAF4EA', // soft banana
    surfaceHover: '#F4EDE2',
    text: '#2A2017',
    textMuted: '#78716c',
    border: '#EDE2D2',
    accent: '#f59e0b',
    accentHover: '#d97706',
    navbar: {
      bg: '#FAF4EA', // soft banana color
      border: '#EDE2D2',
    },
  },

  // Restaurant theme: white background with vibrant orange (used for Restaurant / FrontPage)
  restaurant: {
    primary: '#f05a1f',
    primaryHover: '#d94e17',
    primarySoft: '#fff4ee',
    bg: '#ffffff',
    surface: '#ffffff',
    surfaceHover: '#fff9f5',
    text: '#1f2937',
    textMuted: '#6b7280',
    border: '#f3f4f6',
    accent: '#f05a1f',
    accentHover: '#d94e17',
    navbar: {
      bg: '#e57442', // softer, desaturated warm orange navbar
      border: '#cf6333',
    },
  },
};

export default palette;
