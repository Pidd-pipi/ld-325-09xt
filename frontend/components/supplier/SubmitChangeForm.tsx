'use client';

import { useMemo, useState } from 'react';
import { Send } from 'lucide-react';

import type { Product } from '@/lib/types';
import { money } from '@/lib/utils';
import { Button } from '@/components/ui/button';

interface SubmitChangeFormProps {
  products: Product[];
  submitting: boolean;
  onSubmit: (input: {
    offer_id: number;
    unit_price: number;
    freight: string;
    delivery_days: number;
    stock_status: string;
  }) => Promise<void>;
}

const stockOptions = [
  { value: 'in_stock', label: '有货' },
  { value: 'out_of_stock', label: '缺货' },
  { value: 'discontinued', label: '停产' },
];

// SubmitChangeForm 供应商提交新的单价、运费、货期和库存状态。
export function SubmitChangeForm({ products, submitting, onSubmit }: SubmitChangeFormProps) {
  const withOffers = useMemo(() => products.filter((product) => product.Offers.length > 0), [products]);
  const [productID, setProductID] = useState<number>(withOffers[0]?.ID ?? 0);
  const product = useMemo(
    () => withOffers.find((item) => item.ID === productID) || withOffers[0],
    [withOffers, productID],
  );
  // 演示环境中每个商品的第一条报价归属于 demo 供应商（店铺 1）。
  const offer = product?.Offers.find((item) => item.Supplier.ID === 1) || product?.Offers[0];

  const [price, setPrice] = useState('');
  const [freight, setFreight] = useState('');
  const [days, setDays] = useState('');
  const [stock, setStock] = useState('in_stock');

  if (!product || !offer) {
    return <div className="panel-empty">暂无可维护的报价。</div>;
  }

  const submit = async () => {
    await onSubmit({
      offer_id: offer.ID,
      unit_price: Number(price),
      freight: freight || offer.Freight,
      delivery_days: Number(days),
      stock_status: stock,
    });
    setPrice('');
    setFreight('');
    setDays('');
  };

  const valid = Number(price) > 0 && Number(days) > 0;

  return (
    <div className="calculator change-form">
      <div className="calc-heading"><Send size={16} /> <b>提交报价修改</b></div>
      <label>
        建材报价
        <select value={product.ID} onChange={(event) => setProductID(Number(event.target.value))}>
          {withOffers.map((item) => (
            <option key={item.ID} value={item.ID}>{item.Name} · {item.Offers.find((o) => o.Supplier.ID === 1)?.Supplier.Name || item.Offers[0].Supplier.Name}</option>
          ))}
        </select>
      </label>
      <p className="form-hint">
        当前单价 <strong>{money(offer.UnitPrice)}</strong> · {offer.Freight} · {offer.DeliveryDays} 天交货 · 版本 v{offer.Version}
      </p>
      <label>
        新单价（元）
        <input type="number" min="0.01" step="0.01" value={price} placeholder={String(offer.UnitPrice)} onChange={(event) => setPrice(event.target.value)} />
      </label>
      <label>
        运费说明
        <input type="text" maxLength={100} value={freight} placeholder={offer.Freight} onChange={(event) => setFreight(event.target.value)} />
      </label>
      <div className="change-form-row">
        <label>
          货期（天）
          <input type="number" min="1" max="365" value={days} placeholder={String(offer.DeliveryDays)} onChange={(event) => setDays(event.target.value)} />
        </label>
        <label>
          库存状态
          <select value={stock} onChange={(event) => setStock(event.target.value)}>
            {stockOptions.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}
          </select>
        </label>
      </div>
      <Button disabled={!valid || submitting} onClick={submit}>{submitting ? '提交中…' : '提交审核'}</Button>
      <p className="form-hint subtle">已有待审修改时，新的提交不会入库，需等待本次审核结果。</p>
    </div>
  );
}
