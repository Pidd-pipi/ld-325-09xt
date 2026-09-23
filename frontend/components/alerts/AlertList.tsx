'use client';

import { BellRing, Target, TrendingDown } from 'lucide-react';

import type { AlertView } from '@/lib/types';
import { money } from '@/lib/utils';

const statusMeta: Record<AlertView['status'], { label: string; tone: string }> = {
  active: { label: '监控中', tone: 'good' },
  triggered: { label: '已触发', tone: 'alert' },
  inactive: { label: '已停用', tone: 'neutral' },
};

export function AlertList({ alerts }: { alerts: AlertView[] }) {
  if (alerts.length === 0) {
    return (
      <div className="panel-empty">
        <BellRing size={20} />
        <p>还没有价格提醒。在下方选择材料，设置目标价或降幅后即可订阅。</p>
      </div>
    );
  }
  return (
    <div className="alert-list">
      {alerts.map((alert) => {
        const meta = statusMeta[alert.status];
        return (
          <article className={`alert-card ${alert.status}`} key={alert.id}>
            <div className="alert-card-head">
              <b>{alert.product_name}</b>
              <span className={`badge ${meta.tone}`}>{meta.label}</span>
            </div>
            <div className="alert-price-grid">
              <div>
                <dt><Target size={12} /> 当前最低价</dt>
                <dd className={alert.trigger_price && alert.current_lowest <= alert.trigger_price ? 'hit' : ''}>
                  {alert.current_lowest > 0 ? money(alert.current_lowest) : '暂无有货报价'}
                </dd>
              </div>
              <div>
                <dt><BellRing size={12} /> 触发价</dt>
                <dd>{alert.trigger_price ? money(alert.trigger_price) : '—'}</dd>
              </div>
              <div>
                <dt><TrendingDown size={12} /> 订阅时价</dt>
                <dd>{money(alert.baseline_price)}</dd>
              </div>
            </div>
            <div className="alert-conditions">
              {alert.target_price ? <span>目标价 {money(alert.target_price)}</span> : null}
              {alert.drop_percent ? <span>降幅 ≥ {alert.drop_percent}%</span> : null}
              <span className="alert-time">订阅于 {alert.created_at}</span>
            </div>
            {alert.status === 'triggered' && alert.triggered_at && (
              <p className="alert-hit-note">
                {alert.triggered_at} 以 {money(alert.triggered_price || alert.current_lowest)} 触发，已发送站内提醒。
              </p>
            )}
          </article>
        );
      })}
    </div>
  );
}
