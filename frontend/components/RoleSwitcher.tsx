'use client';

import { useEffect, useState } from 'react';
import { KeyRound, ShieldCheck, Store, UserRound } from 'lucide-react';

import { api, loadTokens, saveToken } from '@/lib/api';
import type { DemoRole } from '@/lib/types';

const roleLabels: Array<{ role: DemoRole; label: string; icon: typeof UserRound }> = [
  { role: 'user', label: '采购用户', icon: UserRound },
  { role: 'supplier', label: '供应商', icon: Store },
  { role: 'admin', label: '平台审核', icon: ShieldCheck },
];

// RoleSwitcher 仅用于演示：为不同角色换取 JWT，驱动页面里的供应商/审核/用户视图。
export function RoleSwitcher({ active, onActive }: { active: DemoRole | null; onActive: (role: DemoRole | null) => void }) {
  const [ready, setReady] = useState<Record<DemoRole, boolean>>({ user: false, supplier: false, admin: false });
  const [busy, setBusy] = useState<DemoRole | null>(null);

  useEffect(() => {
    const sync = () => {
      const tokens = loadTokens();
      setReady({
        user: Boolean(tokens.user),
        supplier: Boolean(tokens.supplier),
        admin: Boolean(tokens.admin),
      });
    };
    sync();
    window.addEventListener('cy-token-changed', sync);
    return () => window.removeEventListener('cy-token-changed', sync);
  }, []);

  const activate = async (role: DemoRole) => {
    setBusy(role);
    try {
      if (!loadTokens()[role]) {
        const token = await api.demoToken(role);
        saveToken(role, token.token);
      }
      onActive(role);
    } finally {
      setBusy(null);
    }
  };

  return (
    <div className="role-switcher">
      <span className="role-hint"><KeyRound size={13} /> 演示身份</span>
      {roleLabels.map(({ role, label, icon: Icon }) => (
        <button
          key={role}
          className={active === role ? 'active' : ''}
          onClick={() => activate(role)}
          disabled={busy !== null}
          title={ready[role] ? '已获取令牌' : '点击获取演示令牌'}
        >
          <Icon size={13} /> {label}
        </button>
      ))}
    </div>
  );
}
