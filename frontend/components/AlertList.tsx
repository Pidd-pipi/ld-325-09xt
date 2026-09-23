'use client';
import { BellRing, BellOff, CheckCircle2, LoaderCircle } from 'lucide-react';
import type { AlertView } from '@/lib/types';
import { alertStatusLabel, cnDateTime, money } from '@/lib/utils';

export function AlertList({ alerts, loading }: { alerts: AlertView[]; loading: boolean }) {
  if (loading) {
    return <div className="loading"><LoaderCircle className="spin" /> 正在加载订阅…</div>;
  }
  if (!alerts.length) {
    return (
      <div className="empty-state">
        <BellOff size={30} />
        <h3>还没有价格提醒</h3>
        <p>在首页选择材料后，设置目标价或降幅。商家改价通过审核时，符合条件的订阅会在这里触发一次。</p>
      </div>
    );
  }
  return (
    <div className="alert-list">
      {alerts.map((alert) => {
        const triggered = alert.status === 'triggered';
        const condition = alert.target_price > 0
          ? `目标价 ${money(alert.target_price)}`
          : `较订阅时降价 ${alert.drop_percent}%`;
        return (
          <article className={triggered ? 'alert-card triggered' : 'alert-card'} key={alert.id}>
            <div className="alert-card-head">
              <div>
                <p className="eyebrow">{alert.brand} · {alert.model}</p>
                <h3>{alert.product_name}</h3>
              </div>
              <span className={`status-pill status-${alert.status}`}>
                {triggered ? <CheckCircle2 size={13} /> : <BellRing size={13} />}
                {alertStatusLabel[alert.status] || alert.status}
              </span>
            </div>
            <dl className="alert-prices">
              <div>
                <dt>当前最低价</dt>
                <dd>{money(alert.current_lowest)}<small>/{alert.unit}</small></dd>
              </div>
              <div>
                <dt>触发价</dt>
                <dd>{money(alert.trigger_price)}</dd>
              </div>
              <div>
                <dt>订阅时低价</dt>
                <dd>{alert.base_price ? money(alert.base_price) : '—'}</dd>
              </div>
            </dl>
            <div className="alert-foot">
              <span className="alert-condition">{condition}</span>
              {triggered ? (
                <span className="alert-fired">已于 {cnDateTime(alert.triggered_at)} 以 {money(alert.triggered_price || 0)} 触发</span>
              ) : (
                <span className="alert-waiting">订阅于 {cnDateTime(alert.created_at)}，等待符合条件的报价</span>
              )}
            </div>
          </article>
        );
      })}
    </div>
  );
}
