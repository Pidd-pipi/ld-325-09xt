import type { DemoSession } from './types';

const SESSION_KEY = 'cybuildprice.session';

export function readSession(): DemoSession | null {
  if (typeof window === 'undefined') return null;
  const raw = window.localStorage.getItem(SESSION_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as DemoSession;
  } catch {
    return null;
  }
}

export function writeSession(session: DemoSession | null) {
  if (typeof window === 'undefined') return;
  if (session) {
    window.localStorage.setItem(SESSION_KEY, JSON.stringify(session));
  } else {
    window.localStorage.removeItem(SESSION_KEY);
  }
  window.dispatchEvent(new CustomEvent('session-change'));
}

export function authToken(): string | undefined {
  return readSession()?.token;
}
