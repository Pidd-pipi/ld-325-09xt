'use client';

import { useMemo, useState } from 'react';
import { BellPlus } from 'lucide-react';

import type { Product } from '@/lib/types';
import { money } from '@/lib/utils';
import { Button } from '@/components/ui/button';

type Mode = 'target' | 'drop';

interface AlertFormProps {
  products: Product[];
  submitting: boolean;
  onSubmit: (input: { product_id: number; target_price?: number; drop_percent?: number }) => Promise<void>;
}

export function AlertForm({ products, submitting, onSubmit }: AlertFormProps) {
  const [productID, setProductID] = useState<number>(products[0]?.ID ?? 0);
  const [mode, setMode] = useState<Mode>('target');
  const [target, setTarget] = useState('');
  const [drop, setDrop] = useState('10');

  const selected = useMemo(
    () => products.find((product) => product.ID === productID) || products[0],
    [products, productID],
  );
  const lowest = selected ? Math.min(...selected.Offers.filter((o) => o.StockStatus === 'in_stock').map((o) => o.UnitPrice)) : 0;
  const preview = mode === 'drop' && lowest > 0 ? lowest * (1 - Number(drop || 0) / 100) : Number(target || 0);

  const submit = async () => {
    const payload = {
      product_id: selected.ID,
      target_price: mode === 'target' ? Number(target) : 0,
      drop_percent: mode === 'drop' ? Number(drop) : 0,
    };
    await onSubmit(payload);
    setTarget('');
  };

  const valid = mode === 'target' ? Number(target) > 0 : Number(drop) > 0;

  return (
    <div className="alert-form calculator">
      <div className="calc-heading"><BellPlus size={16} /> <b>新建价格提醒</b></div>
      <label>
        建材
        <select value={productID} onChange={(event) => setProductID(Number(event.target.value))}>
          {products.map((product) => <option key={product.ID} value={product.ID}>{product.Name}</option>)}
        </select>
      </label>
      {lowest > 0 && <p className="form-hint">当前有货最低价 {money(lowest)}，订阅后只触发一次。</p>}
      <div className="alert-mode" role="group">
        <button className={mode === 'target' ? 'active' : ''} onClick={() => setMode('target')}>目标价</button>
        <button className={mode === 'drop' ? 'active' : ''} onClick={() => setMode('drop')}>降幅百分比</button>
      </div>
      {mode === 'target' ? (
        <label>
          低于多少元提醒
          <input type="number" min="1" step="0.01" value={target} placeholder="例如 360" onChange={(event) => setTarget(event.target.value)} />
        </label>
      ) : (
        <label>
          降价百分比（%）
          <input type="number" min="1" max="100" step="1" value={drop} onChange={(event) => setDrop(event.target.value)} />
        </label>
      )}
      {preview > 0 && <p className="form-hint">预计触发价约 <strong>{money(preview)}</strong></p>}
      <Button disabled={!valid || submitting} onClick={submit}>{submitting ? '提交中…' : '建立提醒'}</Button>
    </div>
  );
}
