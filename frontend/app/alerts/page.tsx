'use client';
import { useEffect, useState } from 'react';
import Link from 'next/link';
import { ArrowLeft } from 'lucide-react';
import { api } from '@/lib/api';
import type { AlertView } from '@/lib/types';
import { AppHeader } from '@/components/AppHeader';
import { AlertList } from '@/components/AlertList';

export default function AlertsPage() {
  const [alerts, setAlerts] = useState<AlertView[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    api.listAlerts()
      .then(setAlerts)
      .catch(() => setError('提醒数据暂时不可用，请确认后端服务已启动。'))
      .finally(() => setLoading(false));
  }, []);

  return (
    <main id="top">
      <AppHeader query="" onQuery={() => undefined} />
      <section className="workspace">
        <div className="workspace-head">
          <div>
            <p className="eyebrow">PRICE ALERTS / 降价订阅</p>
            <h1>价格提醒<em>。</em></h1>
            <p className="workspace-copy">每款材料只保留一条订阅。商家改价通过审核后，达到目标价或设定降幅时触发一次，并记录触发价格与时间。</p>
          </div>
          <Link className="button ghost" href="/#catalog"><ArrowLeft size={15} /> 返回选材料</Link>
        </div>
        {error ? <div className="empty-state"><h3>{error}</h3></div> : <AlertList alerts={alerts} loading={loading} />}
      </section>
      <footer><span>筑价 BUILD PRICE INDEX</span><span>比价、收藏与预算入口在首页继续可用。</span></footer>
    </main>
  );
}
