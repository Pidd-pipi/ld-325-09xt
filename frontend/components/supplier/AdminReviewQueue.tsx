'use client';

import { useCallback, useEffect, useState } from 'react';
import { Check, LoaderCircle, ShieldCheck, X } from 'lucide-react';

import { api } from '@/lib/api';
import type { OfferChangeView } from '@/lib/types';
import { ChangeTable } from '@/components/supplier/ChangeTable';
import { Button } from '@/components/ui/button';

// AdminReviewQueue 管理员审核待审改价。版本冲突时后端会返回 409 并标记失效。
export function AdminReviewQueue() {
  const [rows, setRows] = useState<OfferChangeView[]>([]);
  const [loading, setLoading] = useState(true);
  const [busyID, setBusyID] = useState<number | null>(null);
  const [message, setMessage] = useState('');

  const refresh = useCallback(async () => {
    try {
      setRows(await api.pendingChanges());
    } catch (error) {
      setMessage(error instanceof Error ? error.message : '审核队列加载失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const review = async (row: OfferChangeView, action: 'approve' | 'reject') => {
    setBusyID(row.id);
    setMessage('');
    try {
      await api.reviewChange(row.id, action);
      setMessage(action === 'approve' ? '已通过，新价生效并已核对价格提醒' : '已驳回该修改');
    } catch (error) {
      setMessage(error instanceof Error ? error.message : '审核失败');
    } finally {
      setBusyID(null);
      await refresh();
    }
  };

  return (
    <section id="review" className="workbench-section review-section">
      <div className="section-title">
        <div>
          <p className="eyebrow">ADMIN REVIEW / 报价审核</p>
          <h2>改价先审核，<em>通过后才写进比价。</em></h2>
        </div>
        <div className="catalog-tools">
          <p><ShieldCheck size={14} /> 通过后旧价进入价格历史，并对符合条件的价格提醒触发一次通知；版本已变化的修改会自动失效。</p>
        </div>
      </div>
      {loading ? (
        <div className="loading"><LoaderCircle className="spin" /> 加载待审修改…</div>
      ) : (
        <ChangeTable
          rows={rows}
          emptyText="所有报价修改均已处理完毕。"
          action={(row) => (
            <span className="review-buttons">
              <Button disabled={busyID === row.id} onClick={() => review(row, 'approve')}><Check size={13} /> 通过</Button>
              <button className="ghost-button" disabled={busyID === row.id} onClick={() => review(row, 'reject')}><X size={13} /> 驳回</button>
            </span>
          )}
        />
      )}
      {message && <p className="inline-message" role="status">{message}</p>}
    </section>
  );
}
