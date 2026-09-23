'use client';
import Link from 'next/link';
import { Search, Scale, Bell, Calculator, Store, ClipboardCheck } from 'lucide-react';

export function AppHeader({ query, onQuery }: { query: string; onQuery(value: string): void }) {
  return (
    <header className="site-header">
      <Link className="brand" href="/">
        <span className="brand-mark">筑</span>
        <span>筑价<small>BUILD PRICE INDEX</small></span>
      </Link>
      <nav>
        <a href="/#catalog">找材料</a>
        <a href="/#compare">比报价</a>
        <Link href="/alerts">价格提醒</Link>
        <Link href="/supplier">供应商工作台</Link>
        <Link href="/admin">审核台</Link>
        <a href="/#budget">算预算</a>
      </nav>
      <label className="search">
        <Search size={17} />
        <input value={query} onChange={(event) => onQuery(event.target.value)} placeholder="搜索品牌、型号或建材" />
      </label>
      <div className="header-actions">
        <Link href="/#compare" title="对比清单"><Scale size={18} /></Link>
        <Link href="/alerts" title="价格预警"><Bell size={18} /></Link>
        <Link href="/supplier" title="供应商工作台"><Store size={18} /></Link>
        <Link href="/admin" title="报价审核"><ClipboardCheck size={18} /></Link>
        <a href="#budget" title="预算工具"><Calculator size={18} /></a>
      </div>
    </header>
  );
}
