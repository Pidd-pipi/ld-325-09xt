'use client';
import { useCallback, useEffect, useState } from 'react';
import { Store } from 'lucide-react';
import { api } from '@/lib/api';
import { readSession } from '@/lib/session';
import type { Offer, OfferChangeView } from '@/lib/types';
import { AppHeader } from '@/components/AppHeader';
import { SessionBar } from '@/components/SessionBar';
import { SupplierOffers } from '@/components/SupplierOffers';
import { ChangeForm, type ChangeDraft } from '@/components/ChangeForm';
import { ChangeHistory } from '@/components/ChangeHistory';

export default function SupplierPage() {
  const [isSupplier, setIsSupplier] = useState(false);
  const [offers, setOffers] = useState<Offer[]>([]);
  const [changes, setChanges] = useState<OfferChangeView[]>([]);
  const [loading, setLoading] = useState(false);
  const [selected, setSelected] = useState<Offer | null>(null);
  const [notice, setNotice] = useState('');

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const [offerRows, changeRows] = await Promise.all([api.myOffers(), api.myChanges()]);
      setOffers(offerRows);
      setChanges(changeRows);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    const sync = () => {
      const session = readSession();
      setIsSupplier(session?.role === 'supplier');
      if (session?.role === 'supplier') void refresh();
    };
    sync();
    window.addEventListener('session-change', sync);
    return () => window.removeEventListener('session-change', sync);
  }, [refresh]);

  const submit = async (draft: ChangeDraft) => {
    if (!selected) return;
    await api.submitChange(selected.ID, draft);
    setSelected(null);
    setNotice('修改已提交，等待平台审核。审核期间不能再次提交该报价。');
    window.setTimeout(() => setNotice(''), 4000);
    await refresh();
  };

  const pendingOfferIDs = new Set(changes.filter((item) => item.status === 'pending').map((item) => item.offer_id));

  return (
    <main id="top">
      <AppHeader query="" onQuery={() => undefined} />
      <section className="workspace supplier-workspace">
        <div className="workspace-head">
          <div>
            <p className="eyebrow">SUPPLIER WORKSPACE / 供应商工作台</p>
            <h1>管理报价<em>，</em>跟踪审核<em>。</em></h1>
            <p className="workspace-copy">提交新的单价、运费、货期和库存状态。同一报价存在待审核修改时，新的提交不会入库；审核结果会实时显示。</p>
          </div>
          <SessionBar requireRole="supplier" />
        </div>

        {!isSupplier ? (
          <div className="empty-state">
            <Store size={30} />
            <h3>请先进入演示商家身份</h3>
            <p>右上角的入口会为「筑家优选旗舰店」签发演示令牌，之后即可管理报价并查看审核结果。</p>
          </div>
        ) : (
          <div className="workspace-grid">
            <div>
              <h2 className="panel-title">门店报价</h2>
              <SupplierOffers offers={offers} pendingOfferIDs={pendingOfferIDs} selectedID={selected?.ID ?? null} loading={loading} onSelect={setSelected} />
              {selected && !pendingOfferIDs.has(selected.ID) && (
                <div className="submit-panel">
                  <h3>修改「{selected.Product?.Name}」</h3>
                  <ChangeForm offer={selected} disabled={false} onSubmit={submit} />
                </div>
              )}
              {notice && <div className="toast-inline" role="status">{notice}</div>}
            </div>
            <aside>
              <h2 className="panel-title">审核结果</h2>
              <ChangeHistory changes={changes} />
            </aside>
          </div>
        )}
      </section>
      <footer><span>筑价 BUILD PRICE INDEX</span><span>商家改价经平台审核后才会影响比价与价格提醒。</span></footer>
    </main>
  );
}
