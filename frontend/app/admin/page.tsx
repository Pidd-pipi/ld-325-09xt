'use client';
import { useCallback, useEffect, useState } from 'react';
import { ShieldCheck } from 'lucide-react';
import { api } from '@/lib/api';
import { readSession } from '@/lib/session';
import type { OfferChangeView } from '@/lib/types';
import { AppHeader } from '@/components/AppHeader';
import { SessionBar } from '@/components/SessionBar';
import { ReviewQueue } from '@/components/ReviewQueue';

const tabs = [
  { key: 'pending', label: '待审核' },
  { key: '', label: '全部' },
  { key: 'stale', label: '已失效' },
  { key: 'approved', label: '已通过' },
  { key: 'rejected', label: '已驳回' },
] as const;

export default function AdminPage() {
  const [isAdmin, setIsAdmin] = useState(false);
  const [changes, setChanges] = useState<OfferChangeView[]>([]);
  const [tab, setTab] = useState<string>('pending');
  const [loading, setLoading] = useState(false);

  const refresh = useCallback(async (status: string) => {
    setLoading(true);
    try {
      setChanges(await api.listChanges(status));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    const sync = () => {
      const session = readSession();
      setIsAdmin(session?.role === 'admin');
    };
    sync();
    window.addEventListener('session-change', sync);
    return () => window.removeEventListener('session-change', sync);
  }, []);

  useEffect(() => {
    if (isAdmin) void refresh(tab);
  }, [isAdmin, tab, refresh]);

  return (
    <main id="top">
      <AppHeader query="" onQuery={() => undefined} />
      <section className="workspace admin-workspace">
        <div className="workspace-head">
          <div>
            <p className="eyebrow">REVIEW QUEUE / 报价审核台</p>
            <h1>改价审核<em>。</em></h1>
            <p className="workspace-copy">基于报价版本进行审核：版本已变化时本次修改自动失效并返回冲突；通过后旧价写入历史，达到目标价或降幅的订阅触发一次。</p>
          </div>
          <SessionBar requireRole="admin" />
        </div>

        {!isAdmin ? (
          <div className="empty-state">
            <ShieldCheck size={30} />
            <h3>请先进入演示管理员身份</h3>
            <p>右上角入口会签发管理员演示令牌，随后即可处理供应商提交的报价修改。</p>
          </div>
        ) : (
          <>
            <div className="range-tabs review-tabs" role="group" aria-label="审核状态筛选">
              {tabs.map((item) => (
                <button key={item.key || 'all'} className={tab === item.key ? 'active' : ''} onClick={() => setTab(item.key)}>{item.label}</button>
              ))}
            </div>
            {loading ? <div className="loading">加载中…</div> : <ReviewQueue changes={changes} onReviewed={() => refresh(tab)} />}
          </>
        )}
      </section>
      <footer><span>筑价 BUILD PRICE INDEX</span><span>审核通过的改价才会进入比价、趋势和提醒。</span></footer>
    </main>
  );
}
