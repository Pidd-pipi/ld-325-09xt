'use client';
import { useEffect, useState } from 'react';
import { ShieldCheck, Store, User2, LogOut } from 'lucide-react';
import { api } from '@/lib/api';
import { readSession, writeSession } from '@/lib/session';
import type { DemoSession } from '@/lib/types';

const roleLabel: Record<string, string> = { admin: '管理员', supplier: '供应商', user: '采购用户' };

export function SessionBar({ requireRole }: { requireRole?: 'admin' | 'supplier' }) {
  const [session, setSession] = useState<DemoSession | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    const sync = () => setSession(readSession());
    sync();
    window.addEventListener('session-change', sync);
    return () => window.removeEventListener('session-change', sync);
  }, []);

  const enter = async (role: 'admin' | 'supplier') => {
    setBusy(true);
    setError('');
    try {
      // Demo identities: supplier 1 is the seeded approved store; admin-1 reviews.
      const subject = role === 'supplier' ? '1' : 'admin-1';
      const data = await api.demoToken(subject, role);
      writeSession({ token: data.token, role: data.role, subject: data.subject });
    } catch (err) {
      setError((err as Error).message || '登录失败');
    } finally {
      setBusy(false);
    }
  };

  const missing = requireRole && session?.role !== requireRole;

  return (
    <div className="session-bar">
      <div className="session-identity">
        {session ? (
          <>
            <span className="session-role">
              {session.role === 'admin' ? <ShieldCheck size={15} /> : session.role === 'supplier' ? <Store size={15} /> : <User2 size={15} />}
              {roleLabel[session.role]}
            </span>
            <span className="session-subject">演示身份 #{session.subject}</span>
            <button className="session-logout" onClick={() => writeSession(null)}><LogOut size={14} /> 退出</button>
          </>
        ) : (
          <span className="session-subject">当前以演示采购用户浏览</span>
        )}
      </div>
      <div className="session-actions">
        {missing && requireRole === 'supplier' && <button className="button" disabled={busy} onClick={() => enter('supplier')}><Store size={15} /> 进入演示商家（筑家优选旗舰店）</button>}
        {missing && requireRole === 'admin' && <button className="button" disabled={busy} onClick={() => enter('admin')}><ShieldCheck size={15} /> 进入演示管理员</button>}
        {!missing && session && <button className="session-switch" onClick={() => writeSession(null)}>切换身份</button>}
        {error && <span className="session-error">{error}</span>}
      </div>
    </div>
  );
}
