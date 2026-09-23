import type {
  AlertView,
  ApiEnvelope,
  DemoSession,
  Offer,
  OfferChangeView,
  Product,
  Trend,
} from './types';
import { authToken } from './session';

const API_ROOT = '/api/v1';

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json', ...(init?.headers as Record<string, string> | undefined) };
  const token = authToken();
  if (token) headers.Authorization = `Bearer ${token}`;
  const response = await fetch(`${API_ROOT}${path}`, { ...init, headers });
  const payload = (await response.json()) as ApiEnvelope<T>;
  if (!response.ok || payload.code !== 0) {
    const error = new Error(payload.message) as Error & { status: number; code: number };
    error.status = response.status;
    error.code = payload.code;
    throw error;
  }
  return payload.data;
}

export const api = {
  listProducts: (q = '') => request<{ items: Product[]; total: number }>(`/products?sort=rating&q=${encodeURIComponent(q)}`),
  trend: (id: number, range = '30d') => request<Trend>(`/products/${id}/trend?range=${range}`),
  compare: (ids: number[]) => request<Product[]>('/products/compare', { method: 'POST', body: JSON.stringify({ ids }) }),
  favorite: (productId: number) => request('/favorites', { method: 'POST', body: JSON.stringify({ product_id: productId, folder: '本周采购' }) }),
  budget: (room: string, area: number) =>
    request<{ Estimate: number; Payload: string }>('/budgets', { method: 'POST', body: JSON.stringify({ room_type: room, area }) }),

  // Price subscriptions
  createAlert: (productId: number, targetPrice: number, dropPercent: number) =>
    request<AlertView>('/alerts', { method: 'POST', body: JSON.stringify({ product_id: productId, target_price: targetPrice, drop_percent: dropPercent }) }),
  listAlerts: () => request<AlertView[]>('/alerts'),

  // Demo authentication for supplier/admin views
  demoToken: (subject: string, role: DemoSession['role']) =>
    request<{ token: string; role: DemoSession['role']; subject: string }>('/auth/demo-token', {
      method: 'POST',
      body: JSON.stringify({ subject, role }),
    }),

  // Supplier workspace
  myOffers: () => request<Offer[]>('/supplier/offers'),
  myChanges: (status = '') => request<OfferChangeView[]>(`/supplier/offer-changes${status ? `?status=${status}` : ''}`),
  submitChange: (offerId: number, payload: { unit_price: number; freight: string; delivery_days: number; stock_status: string }) =>
    request<OfferChangeView>(`/supplier/offers/${offerId}/changes`, { method: 'POST', body: JSON.stringify(payload) }),

  // Admin review queue
  listChanges: (status = '') => request<OfferChangeView[]>(`/admin/offer-changes${status ? `?status=${status}` : ''}`),
  reviewChange: (id: number, approve: boolean, note: string) =>
    request<OfferChangeView>(`/admin/offer-changes/${id}/review`, { method: 'POST', body: JSON.stringify({ approve, note }) }),

  // Legacy entry retained by the home page
  alert: (productId: number, target: number) =>
    request<AlertView>('/alerts', { method: 'POST', body: JSON.stringify({ product_id: productId, target_price: target, drop_percent: 10 }) }),
};
