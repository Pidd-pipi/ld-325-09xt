export type Supplier = { ID: number; Name: string; Rating: number; Address: string; Status: string };
export type Offer = { ID: number; UnitPrice: number; MOQ: number; Freight: string; DeliveryDays: number; StockStatus: string; Supplier: Supplier; Version: number };
export type Product = { ID: number; Name: string; Brand: string; Model: string; Unit: string; Thumbnail: string; SalesCount: number; Rating: number; Category: { Name: string }; Offers: Offer[] };
export type ApiEnvelope<T> = { code: number; message: string; data: T };
export type TrendPoint = { Price: number; RecordedAt: string };
export type Trend = { range: string; highest: number; lowest: number; average: number; points: TrendPoint[] };

export type AlertView = {
  id: number;
  product_id: number;
  product_name: string;
  target_price?: number;
  drop_percent?: number;
  baseline_price: number;
  trigger_price?: number;
  current_lowest: number;
  status: 'active' | 'triggered' | 'inactive';
  triggered_price?: number;
  triggered_at?: string;
  created_at: string;
};

export type OfferChangeView = {
  id: number;
  offer_id: number;
  product_id: number;
  product_name: string;
  supplier_id: number;
  supplier_name: string;
  base_unit_price: number;
  new_unit_price: number;
  new_freight: string;
  new_delivery_days: number;
  new_stock_status: string;
  base_version: number;
  current_version: number;
  status: 'pending' | 'approved' | 'rejected' | 'stale';
  review_remark?: string;
  created_at: string;
  reviewed_at?: string;
};

export type NotificationView = {
  id: number;
  title: string;
  content: string;
  product_id: number;
  read: boolean;
  triggered_at: string;
  created_at: string;
};

export type DemoRole = 'admin' | 'supplier' | 'user';
export type DemoToken = { role: DemoRole; user_id: string; token: string; expires_in: number };
