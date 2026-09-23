'use client';
import { useState } from 'react';
import { AlertTriangle, Check, X } from 'lucide-react';
import { api } from '@/lib/api';
import type { OfferChangeView } from '@/lib/types';
import { cnDateTime, changeStatusLabel, money, stockLabel } from '@/lib/utils';

function review(id: number, approve: boolean) {
  return api.reviewChange(id, approve, approve ? '审核通过' : '审核驳回');
}


export function ReviewQueue({ changes, onReviewed }: { changes: OfferChangeView[]; onReviewed: () => Promise<void> }) {
  const [busyID, setBusyID] = useState<number | null>(null);
  const [conflict, setConflict] = useState<{ id: number; message: string } | null>(null);

  const act = async (change: OfferChangeView, approve: boolean) => {
    setBusyID(change.id);
    setConflict(null);
    try {
      await review(change.id, approve);
      await onReviewed();
    } catch (err) {
      const status = (err as { status?: number }).status;
      if (status === 409) {
        setConflict({ id: change.id, message: (err as Error).message });
        await onReviewed();
      } else {
        setConflict({ id: change.id, message: (err as Error).message || '操作失败' });
      }
    } finally {
      setBusyID(null);
    }
  };

  return (
    <div className="review-list">
      {conflict && (
        <div className="conflict-banner" role="alert">
          <AlertTriangle size={17} /> <span>{conflict.message} 该修改已标记为失效。</span>
        </div>
      )}
      {changes.length === 0 && (
        <div className="empty-state"><h3>没有待审核的报价修改</h3><p>供应商提交新单价、运费、货期或库存状态后，会进入这个队列。</p></div>
      )}
      {changes.map((change) => {
        const priceUp = change.new_unit_price > change.current_price;
        return (
          <article key={change.id} className="review-card">
            <div className="review-head">
              <div>
                <p className="eyebrow">{change.supplier_name} · 基于报价 v{change.base_version} / 当前 v{change.offer_version}</p>
                <h3>{change.product_name || `报价 #${change.offer_id}`}</h3>
              </div>
              <span className={`status-pill status-${change.status}`}>{changeStatusLabel[change.status] || change.status}</span>
            </div>
            <div className="review-table">
              <div className="review-col"><span>字段</span><b>现行</b><b>申请修改</b></div>
              <div className="review-col"><span>单价</span><b>{money(change.current_price)}</b><b className={priceUp ? 'price-up' : 'price-down'}>{money(change.new_unit_price)}</b></div>
              <div className="review-col"><span>货期</span><b>{change.current_delivery_days} 天</b><b>{change.new_delivery_days} 天</b></div>
              <div className="review-col"><span>库存</span><b>{stockLabel[change.current_stock_status]}</b><b>{stockLabel[change.new_stock_status]}</b></div>
              <div className="review-col wide"><span>运费</span><b>{change.current_freight}</b><b>{change.new_freight}</b></div>
            </div>
            <div className="review-foot">
              <span>提交于 {cnDateTime(change.created_at)}</span>
              {change.status === 'pending' && (
                <div className="review-actions">
                  <button className="button reject" disabled={busyID === change.id} onClick={() => act(change, false)}><X size={15} /> 驳回</button>
                  <button className="button approve" disabled={busyID === change.id} onClick={() => act(change, true)}><Check size={15} /> 通过</button>
                </div>
              )}
              {(change.status === 'stale' || change.status === 'rejected' || change.status === 'approved') && (
                <span className="review-note">{cnDateTime(change.reviewed_at)} {change.review_note ? `· ${change.review_note}` : ''}</span>
              )}
            </div>
          </article>
        );
      })}
    </div>
  );
}
