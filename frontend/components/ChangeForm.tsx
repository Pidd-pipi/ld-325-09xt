'use client';
import { useState } from 'react';
import type { Offer } from '@/lib/types';
import { money, stockLabel } from '@/lib/utils';

export interface ChangeDraft {
  unit_price: number;
  freight: string;
  delivery_days: number;
  stock_status: string;
}

export function ChangeForm({
  offer,
  disabled,
  onSubmit,
}: {
  offer: Offer;
  disabled: boolean;
  onSubmit: (draft: ChangeDraft) => Promise<void>;
}) {
  const [unitPrice, setUnitPrice] = useState(String(offer.UnitPrice));
  const [freight, setFreight] = useState(offer.Freight);
  const [deliveryDays, setDeliveryDays] = useState(String(offer.DeliveryDays));
  const [stockStatus, setStockStatus] = useState(offer.StockStatus);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  const submit = async () => {
    const price = Number(unitPrice);
    const days = Number(deliveryDays);
    if (!(price > 0)) {
      setError('请输入有效的单价');
      return;
    }
    if (!freight.trim()) {
      setError('请填写运费说明');
      return;
    }
    setBusy(true);
    setError('');
    try {
      await onSubmit({ unit_price: price, freight: freight.trim(), delivery_days: days, stock_status: stockStatus });
    } catch (err) {
      setError((err as Error).message || '提交失败');
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="change-form">
      <div className="form-grid">
        <label>单价（元/{offer.Product?.Unit || '件'})
          <input type="number" min="0.01" step="0.01" value={unitPrice} onChange={(e) => setUnitPrice(e.target.value)} />
        </label>
        <label>货期（天）
          <input type="number" min="0" max="365" value={deliveryDays} onChange={(e) => setDeliveryDays(e.target.value)} />
        </label>
        <label>库存状态
          <select value={stockStatus} onChange={(e) => setStockStatus(e.target.value)}>
            <option value="in_stock">有货</option>
            <option value="out_of_stock">缺货</option>
            <option value="discontinued">停产</option>
          </select>
        </label>
        <label className="form-wide">运费说明
          <input value={freight} maxLength={200} onChange={(e) => setFreight(e.target.value)} />
        </label>
      </div>
      <div className="form-foot">
        <span className="form-hint">现行单价 {money(offer.UnitPrice)} · {stockLabel[offer.StockStatus]} · 报价版本 v{offer.Version}</span>
        <button className="button" disabled={busy || disabled} onClick={submit}>{busy ? '提交中…' : '提交修改审核'}</button>
      </div>
      {error && <p className="form-error">{error}</p>}
    </div>
  );
}
