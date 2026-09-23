export type Supplier = { ID: number; Name: string; Rating: number; Address: string; Status: string };
export type Offer = {
  ID: number;
  ProductID: number;
  SupplierID: number;
  UnitPrice: number;
  MOQ: number;
  Freight: string;
  DeliveryDays: number;
  StockStatus: string;
  Version: number;
  Supplier: Supplier;
  Product?: Product;
};
export type Product = {
  ID: number;
  Name: string;
  Brand: string;
  Model: string;
  Unit: string;
  Thumbnail: string;
  SalesCount: number;
  Rating: number;
  Category: { Name: string };
  Offers: Offer[];
};
export type ApiEnvelope<T> = { code: number; message: string; data: T };
export type TrendPoint = { Price: number; RecordedAt: string };
export type Trend = { range: string; highest: number; lowest: number; average: number; points: TrendPoint[] };

export type AlertView = {
  id: number;
  product_id: number;
  product_name: string;
  brand: string;
  model: string;
  unit: string;
  base_price: number;
  target_price: number;
  drop_percent: number;
  trigger_price: number;
  current_lowest: number;
  status: 'active' | 'triggered' | 'inactive' | string;
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
  current_price: number;
  new_unit_price: number;
  current_freight: string;
  new_freight: string;
  current_delivery_days: number;
  new_delivery_days: number;
  current_stock_status: string;
  new_stock_status: string;
  base_version: number;
  offer_version: number;
  status: 'pending' | 'approved' | 'rejected' | 'stale' | string;
  review_note: string;
  created_at: string;
  reviewed_at?: string;
};

export type DemoSession = { token: string; role: 'admin' | 'supplier' | 'user'; subject: string };
