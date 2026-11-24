export type UserRestaurant = {
  slug: string;
  name: string;
  role: string;
  status: string;
};

// Replace this boundary with the restaurant membership API when it is available.
export async function getMyRestaurants(): Promise<UserRestaurant[]> {
  if (import.meta.env.TEST_ENV === 'true' || import.meta.env.TEST_ENV === '1') {
    return [
      { slug: 'joao-burger', name: 'João Burger', role: 'Administrador', status: 'Ativo' },
      { slug: 'cantina-bella', name: 'Cantina Bella', role: 'Administrador', status: 'Ativo' },
      { slug: 'sabor-da-casa', name: 'Sabor da Casa', role: 'Funcionário', status: 'Ativo' }
    ];
  }
  return [];
}
