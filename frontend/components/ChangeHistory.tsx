'use client';
import { cnDateTime, changeStatusLabel, money, stockLabel } from '@/lib/utils';
import type { OfferChangeView } from '@/lib/types';

export function ChangeHistory({ changes }: { changes: OfferChangeView[] }) {
  if (!changes.length) {
    return <p className="muted-note">还没有修改记录。提交单价、运费、货期或库存变更后，审核结果会出现在这里。</p>;
  }
  return (
    <div className="change-history">
      {changes.map((change) => (
        <article key={change.id} className={`change-row status-${change.status}`}>
          <div className="change-row-head">
            <b>{change.product_name || `报价 #${change.offer_id}`}</b>
            <span className={`status-pill status-${change.status}`}>{changeStatusLabel[change.status] || change.status}</span>
          </div>
          <div className="change-compare">
            <span className="change-old">{money(change.current_price)} · {stockLabel[change.current_stock_status]} · {change.current_delivery_days} 天</span>
            <span aria-hidden="true">→</span>
            <span className="change-new">{money(change.new_unit_price)} · {stockLabel[change.new_stock_status]} · {change.new_delivery_days} 天</span>
          </div>
          <p className="change-freight">运费：{change.current_freight} → {change.new_freight}</p>
          <footer>
            <span>提交于 {cnDateTime(change.created_at)} · 基于 v{change.base_version}</span>
            <span>{change.status === 'pending' ? `当前报价 v${change.offer_version}` : `${changeStatusLabel[change.status]}于 ${cnDateTime(change.reviewed_at)}`}{change.review_note ? ` · ${change.review_note}` : ''}</span>
          </footer>
        </article>
      ))}
    </div>
  );
}
