'use client';
import { Calculator, Scale, Search } from 'lucide-react';
import { RoleSwitcher } from './RoleSwitcher';
import type { DemoRole } from '@/lib/types';

export function AppHeader({
  query,
  onQuery,
  role,
  onRole,
}: {
  query: string;
  onQuery(value: string): void;
  role: DemoRole | null;
  onRole(role: DemoRole | null): void;
}) {
  return (
    <header className="site-header">
      <a className="brand" href="#top"><span className="brand-mark">筑</span><span>筑价<small>BUILD PRICE INDEX</small></span></a>
      <nav>
        <a href="#catalog">找材料</a>
        <a href="#compare"><Scale size={12} /> 比报价</a>
        <a href="#alerts">价格提醒</a>
        <a href="#supplier">供应商</a>
        <a href="#budget"><Calculator size={12} /> 算预算</a>
      </nav>
      <label className="search"><Search size={17} /><input value={query} onChange={(event) => onQuery(event.target.value)} placeholder="搜索品牌、型号或建材" /></label>
      <div className="header-actions">
        <RoleSwitcher active={role} onActive={onRole} />
      </div>
    </header>
  );
}
