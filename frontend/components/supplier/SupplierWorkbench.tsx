'use client';

import { useCallback, useEffect, useState } from 'react';
import { LoaderCircle, Store } from 'lucide-react';

import { api, loadTokens } from '@/lib/api';
import type { OfferChangeView, Product } from '@/lib/types';
import { ChangeTable } from './ChangeTable';
import { SubmitChangeForm } from './SubmitChangeForm';

// SupplierWorkbench 供应商提交改价并查看审核结果（通过 / 驳回 / 失效）。
export function SupplierWorkbench({ products }: { products: Product[] }) {
  const [rows, setRows] = useState<OfferChangeView[]>([]);
  const [loading, setLoading] = useState(true);
  const [authenticated, setAuthenticated] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [message, setMessage] = useState('');

  const refresh = useCallback(async () => {
    if (!loadTokens().supplier) {
      setAuthenticated(false);
      setLoading(false);
      return;
    }
    setAuthenticated(true);
    try {
      setRows(await api.supplierChanges());
    } catch (error) {
      setMessage(error instanceof Error ? error.message : '审核结果加载失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh();
    const sync = () => refresh();
    window.addEventListener('cy-token-changed', sync);
    return () => window.removeEventListener('cy-token-changed', sync);
  }, [refresh]);

  const submit = async (input: {
    offer_id: number;
    unit_price: number;
    freight: string;
    delivery_days: number;
    stock_status: string;
  }) => {
    setSubmitting(true);
    try {
      await api.submitChange(input);
      setMessage('报价修改已提交，等待平台审核');
      await refresh();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : '提交失败');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <section id="supplier" className="workbench-section">
      <div className="section-title">
        <div>
          <p className="eyebrow">SUPPLIER WORKBENCH / 供应商工作台</p>
          <h2>提交改价，<em>审核结果公开可查。</em></h2>
        </div>
        <div className="catalog-tools">
          <p><Store size={14} /> 演示身份：筑家优选旗舰店（supplier-1）。单价、运费、货期与库存状态提交后进入审核。</p>
        </div>
      </div>
      {!authenticated ? (
        <div className="panel-empty">
          <Store size={20} />
          <p>请在页面顶部切换到「供应商」演示身份，即可提交新单价、运费、货期和库存状态，并查看审核结果。</p>
        </div>
      ) : (
        <div className="workbench-layout">
          <SubmitChangeForm products={products} submitting={submitting} onSubmit={submit} />
          <div className="change-panel">
            <h3>我的修改与审核结果</h3>
            {loading ? (
              <div className="loading"><LoaderCircle className="spin" /> 加载中…</div>
            ) : (
              <ChangeTable rows={rows} emptyText="还没有提交过报价修改。" />
            )}
          </div>
        </div>
      )}
      {message && <p className="inline-message" role="status">{message}</p>}
    </section>
  );
}
