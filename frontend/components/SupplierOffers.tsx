'use client';
import { LoaderCircle } from 'lucide-react';
import type { Offer } from '@/lib/types';
import { money, stockLabel } from '@/lib/utils';

export function SupplierOffers({
  offers,
  pendingOfferIDs,
  selectedID,
  loading,
  onSelect,
}: {
  offers: Offer[];
  pendingOfferIDs: Set<number>;
  selectedID: number | null;
  loading: boolean;
  onSelect: (offer: Offer) => void;
}) {
  if (loading) {
    return <div className="loading"><LoaderCircle className="spin" /> 正在读取门店报价…</div>;
  }
  if (!offers.length) {
    return <div className="empty-state"><h3>当前商家暂无可管理的报价</h3><p>演示身份「筑家优选旗舰店」在平台已审核的报价会显示在这里。</p></div>;
  }
  return (
    <div className="offer-manage-list">
      {offers.map((offer) => {
        const pending = pendingOfferIDs.has(offer.ID);
        const selected = selectedID === offer.ID;
        return (
          <article key={offer.ID} className={selected ? 'offer-manage selected' : 'offer-manage'}>
            <div>
              <p className="eyebrow">{stockLabel[offer.StockStatus]} · v{offer.Version}</p>
              <h3>{offer.Product?.Name || `报价 #${offer.ID}`}</h3>
              <p>{offer.Freight} · {offer.DeliveryDays} 天交货</p>
            </div>
            <strong>{money(offer.UnitPrice)}</strong>
            {pending ? (
              <span className="status-pill status-pending">有待审核修改</span>
            ) : (
              <button className="button ghost" onClick={() => onSelect(offer)}>修改报价</button>
            )}
          </article>
        );
      })}
    </div>
  );
}
