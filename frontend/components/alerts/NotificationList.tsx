'use client';

import { Inbox } from 'lucide-react';

import type { NotificationView } from '@/lib/types';

export function NotificationList({ notifications }: { notifications: NotificationView[] }) {
  if (notifications.length === 0) {
    return (
      <div className="panel-empty subtle">
        <Inbox size={18} />
        <p>暂无站内提醒。符合目标价或降幅的报价审核通过后，会在这里通知你。</p>
      </div>
    );
  }
  return (
    <ul className="notification-list">
      {notifications.map((item) => (
        <li key={item.id} className={item.read ? 'read' : 'unread'}>
          <b>{item.title}</b>
          <p>{item.content}</p>
          <time>{item.triggered_at}</time>
        </li>
      ))}
    </ul>
  );
}
