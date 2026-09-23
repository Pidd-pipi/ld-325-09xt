export const money = (value: number) =>
  new Intl.NumberFormat('zh-CN', { style: 'currency', currency: 'CNY', maximumFractionDigits: 0 }).format(value);

export const cnDate = (value: string) =>
  new Intl.DateTimeFormat('zh-CN', { month: 'numeric', day: 'numeric' }).format(new Date(value));

export const cnDateTime = (value?: string) =>
  value
    ? new Intl.DateTimeFormat('zh-CN', { month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit' }).format(new Date(value))
    : '—';

export const stockLabel: Record<string, string> = {
  in_stock: '有货',
  out_of_stock: '缺货',
  discontinued: '停产',
};

export const changeStatusLabel: Record<string, string> = {
  pending: '待审核',
  approved: '已通过',
  rejected: '已驳回',
  stale: '已失效',
};

export const alertStatusLabel: Record<string, string> = {
  active: '监控中',
  triggered: '已触发',
  inactive: '已停用',
};
