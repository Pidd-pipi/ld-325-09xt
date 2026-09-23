'use client';

import { useCallback, useEffect, useState } from 'react';
import { LoaderCircle } from 'lucide-react';

import { api } from '@/lib/api';
import type { AlertView, NotificationView, Product } from '@/lib/types';
import { AlertList } from './AlertList';
import { AlertForm } from './AlertForm';
import { NotificationList } from './NotificationList';

// AlertCenter 提醒页：展示当前最低价、触发价和状态，并集中管理订阅与站内消息。
export function AlertCenter({ products }: { products: Product[] }) {
  const [alerts, setAlerts] = useState<AlertView[]>([]);
  const [notifications, setNotifications] = useState<NotificationView[]>([]);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [message, setMessage] = useState('');

  const refresh = useCallback(async () => {
    try {
      const [alertRows, messages] = await Promise.all([api.alerts(), api.notifications()]);
      setAlerts(alertRows);
      setNotifications(messages);
    } catch (error) {
      setMessage(error instanceof Error ? error.message : '提醒数据加载失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const createAlert = async (input: { product_id: number; target_price?: number; drop_percent?: number }) => {
    setSubmitting(true);
    try {
      await api.createAlert(input);
      setMessage('价格提醒已建立');
      await refresh();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : '建立提醒失败');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <section id="alerts" className="alerts-section">
      <div className="section-title">
        <div>
          <p className="eyebrow">PRICE ALERT / 降价提醒</p>
          <h2>价格到位，<em>站内第一时间通知。</em></h2>
        </div>
        <div className="catalog-tools">
          <p>同一材料只保留一条生效订阅；报价修改经平台审核通过后，符合目标价或降幅时触发一次。</p>
        </div>
      </div>

      <div className="alerts-layout">
        <div>
          {loading ? (
            <div className="loading"><LoaderCircle className="spin" /> 正在加载订阅…</div>
          ) : (
            <AlertList alerts={alerts} />
          )}
          <div className="notification-block">
            <h3>站内提醒</h3>
            <NotificationList notifications={notifications} />
          </div>
        </div>
        <AlertForm products={products} submitting={submitting} onSubmit={createAlert} />
      </div>
      {message && <p className="inline-message" role="status">{message}</p>}
    </section>
  );
}
