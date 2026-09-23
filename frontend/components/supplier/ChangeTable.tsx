'use client';

import type { ReactNode } from 'react';
import { ClipboardList } from 'lucide-react';

import type { OfferChangeView } from '@/lib/types';
import { money } from '@/lib/utils';

export const changeStatusMeta: Record<OfferChangeView['status'], { label: string; tone: string }> = {
  pending: { label: '待审核', tone: 'neutral' },
  approved: { label: '已通过', tone: 'good' },
  rejected: { label: '已驳回', tone: 'alert' },
  stale: { label: '已失效', tone: 'alert' },
};

const stockText: Record<string, string> = {
  in_stock: '有货',
  out_of_stock: '缺货',
  discontinued: '停产',
};

interface ChangeTableProps {
  rows: OfferChangeView[];
  action?: (row: OfferChangeView) => ReactNode;
  emptyText?: string;
}

export function ChangeTable({ rows, action, emptyText = '暂无修改记录。' }: ChangeTableProps) {
  if (rows.length === 0) {
    return (
      <div className="panel-empty">
        <ClipboardList size={18} />
        <p>{emptyText}</p>
      </div>
    );
  }
  return (
    <div className="change-table">
      <div className="change-row change-row-head">
        <span>建材 / 商家</span>
        <span>价格变更</span>
        <span>交付与库存</span>
        <span>状态</span>
        {action ? <span>操作</span> : null}
      </div>
      {rows.map((row) => {
        const meta = changeStatusMeta[row.status];
        const delta = row.new_unit_price - row.base_unit_price;
        return (
          <div className="change-row" key={row.id}>
            <span className="change-product">
              <b>{row.product_name}</b>
              <small>{row.supplier_name}</small>
              <small className="change-time">提交于 {row.created_at}</small>
            </span>
            <span className="change-price">
              <s>{money(row.base_unit_price)}</s>
              <strong className={delta < 0 ? 'down' : delta > 0 ? 'up' : ''}>{money(row.new_unit_price)}</strong>
              <small>
                {delta < 0 ? '↓' : delta > 0 ? '↑' : '='} {Math.abs(delta).toFixed(2)} 元
              </small>
            </span>
            <span className="change-delivery">
              <small>{row.new_freight}</small>
              <small>{row.new_delivery_days} 天交货 · {stockText[row.new_stock_status] || row.new_stock_status}</small>
              <small>基准 v{row.base_version} / 当前 v{row.current_version}</small>
            </span>
            <span className="change-status">
              <span className={`badge ${meta.tone}`}>{meta.label}</span>
              {row.reviewed_at ? <small>{row.reviewed_at} 审核</small> : null}
              {row.review_remark ? <small className="remark">备注：{row.review_remark}</small> : null}
            </span>
            {action ? <span className="change-actions">{action(row)}</span> : null}
          </div>
        );
      })}
    </div>
  );
}
