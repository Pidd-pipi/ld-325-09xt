import type {
  AlertView,
  ApiEnvelope,
  DemoRole,
  DemoToken,
  NotificationView,
  OfferChangeView,
  Product,
  Trend,
} from './types';

const API_ROOT = '/api/v1';
const TOKEN_STORAGE_KEY = 'cybuildprice.tokens';

type TokenMap = Partial<Record<DemoRole, string>>;

export function loadTokens(): TokenMap {
  if (typeof window === 'undefined') return {};
  try {
    return JSON.parse(window.localStorage.getItem(TOKEN_STORAGE_KEY) || '{}') as TokenMap;
  } catch {
    return {};
  }
}

export function saveToken(role: DemoRole, token: string): TokenMap {
  const tokens = { ...loadTokens(), [role]: token };
  window.localStorage.setItem(TOKEN_STORAGE_KEY, JSON.stringify(tokens));
  window.dispatchEvent(new CustomEvent('cy-token-changed'));
  return tokens;
}

export function clearTokens(): void {
  window.localStorage.removeItem(TOKEN_STORAGE_KEY);
  window.dispatchEvent(new CustomEvent('cy-token-changed'));
}

function authHeader(role?: DemoRole): HeadersInit {
  if (!role) return {};
  const token = loadTokens()[role];
  return token ? { Authorization: `Bearer ${token}` } : {};
}

async function request<T>(path: string, init?: RequestInit, role?: DemoRole): Promise<T> {
  const response = await fetch(`${API_ROOT}${path}`, {
    headers: { 'Content-Type': 'application/json', ...authHeader(role), ...(init?.headers || {}) },
    ...init,
  });
  const payload = (await response.json()) as ApiEnvelope<T>;
  if (!response.ok || payload.code !== 0) {
    const error = new Error(payload.message) as Error & { code?: number; status?: number };
    error.code = payload.code;
    error.status = response.status;
    throw error;
  }
  return payload.data;
}

export const api = {
  listProducts: (q = '') =>
    request<{ items: Product[]; total: number }>(`/products?sort=rating&q=${encodeURIComponent(q)}`),
  product: (id: number) => request<Product>(`/products/${id}`),
  trend: (id: number, range = '30d') => request<Trend>(`/products/${id}/trend?range=${range}`),
  compare: (ids: number[]) => request<Product[]>('/products/compare', { method: 'POST', body: JSON.stringify({ ids }) }),
  favorite: (productId: number) =>
    request('/favorites', { method: 'POST', body: JSON.stringify({ product_id: productId, folder: '本周采购' }) }),
  budget: (room: string, area: number) =>
    request<{ Estimate: number; Payload: string }>('/budgets', {
      method: 'POST',
      body: JSON.stringify({ room_type: room, area }),
    }),

  // 价格提醒
  alert: (productId: number, target: number) =>
    request<AlertView>('/alerts', {
      method: 'POST',
      body: JSON.stringify({ product_id: productId, target_price: target, drop_percent: 0 }),
    }),
  createAlert: (input: { product_id: number; target_price?: number; drop_percent?: number }) =>
    request<AlertView>('/alerts', { method: 'POST', body: JSON.stringify(input) }),
  alerts: () => request<AlertView[]>('/alerts', undefined, 'user'),
  notifications: () => request<NotificationView[]>('/notifications', undefined, 'user'),

  // 供应商改价
  demoToken: (role: DemoRole) =>
    request<DemoToken>('/auth/demo-token', { method: 'POST', body: JSON.stringify({ role }) }),
  supplierChanges: () => request<OfferChangeView[]>('/supplier/offer-changes', undefined, 'supplier'),
  submitChange: (input: {
    offer_id: number;
    unit_price: number;
    freight: string;
    delivery_days: number;
    stock_status: string;
  }) =>
    request<OfferChangeView>('/supplier/offer-changes', { method: 'POST', body: JSON.stringify(input) }, 'supplier'),

  // 管理员审核
  pendingChanges: () => request<OfferChangeView[]>('/admin/offer-changes', undefined, 'admin'),
  reviewChange: (id: number, action: 'approve' | 'reject', remark = '') =>
    request<OfferChangeView>(`/admin/offer-changes/${id}/review`, {
      method: 'POST',
      body: JSON.stringify({ action, remark }),
    }, 'admin'),
};
