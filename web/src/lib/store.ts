import { create } from 'zustand';

const initial = JSON.parse(localStorage.getItem('user') || 'null');

export const useStore = create((set) => ({
  user: initial,
  token: localStorage.getItem('token') || null,
  setUser: (user, token) => {
    localStorage.setItem('user', JSON.stringify(user));
    localStorage.setItem('token', token);
    set({ user, token });
  },
  logout: () => {
    localStorage.removeItem('user');
    localStorage.removeItem('token');
    set({ user: null, token: null });
  },
}));