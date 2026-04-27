import { createContext, useContext, useState, ReactNode, useEffect } from "react";
import { fetchDriverProfile, transformProfileData } from "./api/profile";

const TEST_DRIVER_ID = 200000001;

export type OrderStatus =
  | "pending"
  | "accepted"
  | "arrived"
  | "ongoing"
  | "toPay"
  | "completed"
  | "cancelled";

export interface Order {
  id: string;
  passengerName: string;
  passengerPhone: string;
  from: string;
  to: string;
  distanceKm: number;
  estMinutes: number;
  price: number;
  originalPrice?: number;
  couponId?: string;
  couponDiscount?: number;
  carType: string;
  status: OrderStatus;
  createdAt: number;
  note?: string;
  isPrebook?: boolean;
  prebookTime?: string;
  driverId?: string;
  driverName?: string;
  driverPhone?: string;
  driverPlate?: string;
  driverCar?: string;
  driverRating?: number;
  paymentMethod?: string;
  rating?: number;
  ratingComment?: string;
  ratingTags?: string[];
  coinsEarned?: number;
}

export interface Driver {
  id: string;
  name: string;
  phone: string;
  plate: string;
  car: string;
  rating: number;
  online: boolean;
  totalOrders: number;
  todayEarnings: number;
  status: "idle" | "busy" | "offline";
}

export interface Complaint {
  id: string;
  orderId: string;
  from: "passenger" | "driver";
  content: string;
  status: "open" | "resolved";
  createdAt: number;
}

interface Store {
  orders: Order[];
  drivers: Driver[];
  complaints: Complaint[];
  currentDriverId: string;
  createOrder: (o: Omit<Order, "id" | "createdAt" | "status">) => string;
  acceptOrder: (orderId: string, driverId: string) => void;
  updateOrderStatus: (orderId: string, status: OrderStatus) => void;
  cancelOrder: (orderId: string) => void;
  setDriverOnline: (driverId: string, online: boolean) => void;
}

const StoreCtx = createContext<Store | null>(null);

const initialDrivers: Driver[] = [
  { id: "D001", name: "王师傅", phone: "138****2341", plate: "苏N·8F23K", car: "轩逸 · 银色", rating: 4.92, online: true, totalOrders: 2341, todayEarnings: 328, status: "idle" },
  { id: "D002", name: "李师傅", phone: "139****6612", plate: "苏N·1K88P", car: "朗逸 · 白色", rating: 4.88, online: true, totalOrders: 1892, todayEarnings: 265, status: "idle" },
  { id: "D003", name: "张师傅", phone: "136****9900", plate: "苏N·2M55Q", car: "捷达 · 黑色", rating: 4.76, online: false, totalOrders: 985, todayEarnings: 0, status: "offline" },
  { id: "D004", name: "赵师傅", phone: "135****4411", plate: "苏N·6R23L", car: "卡罗拉 · 白色", rating: 4.95, online: true, totalOrders: 3102, todayEarnings: 412, status: "idle" },
];

const initialOrders: Order[] = [
  { id: "20260423001", passengerName: "我", passengerPhone: "198****2059", from: "宿迁职业技术学院8号楼南侧", to: "宿迁万达广场", distanceKm: 6.2, estMinutes: 18, price: 11.2, originalPrice: 17.2, couponDiscount: 6, carType: "小猪特价", status: "completed", createdAt: Date.now() - 3600000 * 3, driverId: "D001", driverName: "王师傅", driverPlate: "苏N·8F23K", driverCar: "轩逸 · 银色", driverRating: 4.92, paymentMethod: "微信支付", rating: 5, ratingComment: "师傅很准时", ratingTags: ["准时准点", "车内整洁"], coinsEarned: 22 },
  { id: "20260422002", passengerName: "我", passengerPhone: "198****2059", from: "宿迁高铁站", to: "宿迁职业技术学院8号楼南侧", distanceKm: 12.5, estMinutes: 28, price: 21.8, originalPrice: 26.8, couponDiscount: 5, carType: "小猪特价", status: "completed", createdAt: Date.now() - 86400000, driverId: "D002", driverName: "李师傅", driverPlate: "苏N·1K88P", driverCar: "朗逸 · 白色", driverRating: 4.88, paymentMethod: "微信支付", rating: 5, ratingComment: "路线很熟悉", coinsEarned: 43 },
  { id: "20260420003", passengerName: "我", passengerPhone: "198****2059", from: "宿迁人民医院", to: "宿迁中央商务区", distanceKm: 4.8, estMinutes: 14, price: 9.2, carType: "小猪特价", status: "cancelled", createdAt: Date.now() - 86400000 * 3 },
  { id: "20260423004", passengerName: "用户8821", passengerPhone: "138****0002", from: "宿迁万达广场", to: "宿迁高铁站", distanceKm: 8.2, estMinutes: 22, price: 24.0, carType: "特惠快车", status: "ongoing", createdAt: Date.now() - 600000, driverId: "D002", driverName: "李师傅", driverPlate: "苏N·1K88P", driverCar: "朗逸 · 白色", driverRating: 4.88 },
  { id: "20260423005", passengerName: "用户1109", passengerPhone: "138****0003", from: "宿迁市政府广场", to: "宿迁学院北门", distanceKm: 12.8, estMinutes: 28, price: 36.5, carType: "舒适型", status: "pending", createdAt: Date.now() - 120000 },
];

const initialComplaints: Complaint[] = [
  { id: "C001", orderId: "20260423004", from: "passenger", content: "司机绕路了2公里", status: "open", createdAt: Date.now() - 7200000 },
  { id: "C002", orderId: "20260422002", from: "passenger", content: "车内有异味，体验不佳", status: "resolved", createdAt: Date.now() - 86400000 },
];

export function StoreProvider({ children }: { children: ReactNode }) {
  const [orders, setOrders] = useState<Order[]>(initialOrders);
  const [drivers, setDrivers] = useState<Driver[]>(initialDrivers);
  const [complaints, setComplaints] = useState<Complaint[]>(initialComplaints);
  const [currentDriverId] = useState("D001");
  const [profileLoaded, setProfileLoaded] = useState(false);

  // 初始化时从后端加载司机 Profile
  useEffect(() => {
    async function loadProfile() {
      const data = await fetchDriverProfile(TEST_DRIVER_ID);
      if (data) {
        const driverData = transformProfileData(data, "D001");
        setDrivers((prev) =>
          prev.map((d) => (d.id === "D001" ? { ...d, ...driverData } : d))
        );
        console.log("Profile loaded from backend:", driverData);
      }
      setProfileLoaded(true);
    }
    loadProfile();
  }, []);

  useEffect(() => {
    const t = setInterval(() => {
      setOrders((prev) => prev.map((o) => {
        const age = Date.now() - o.createdAt;
        if (o.status === "accepted" && age > 15000) return { ...o, status: "arrived" };
        if (o.status === "arrived" && age > 30000) return { ...o, status: "ongoing" };
        return o;
      }));
    }, 2000);
    return () => clearInterval(t);
  }, []);

  const createOrder: Store["createOrder"] = (o) => {
    const id = "2026" + String(Date.now()).slice(-8);
    setOrders((p) => [{ ...o, id, createdAt: Date.now(), status: "pending" }, ...p]);
    return id;
  };

  const acceptOrder: Store["acceptOrder"] = (orderId, driverId) => {
    const d = drivers.find((x) => x.id === driverId);
    if (!d) return;
    setOrders((p) => p.map((o) => o.id === orderId ? {
      ...o, status: "accepted", driverId, driverName: d.name,
      driverPhone: d.phone, driverPlate: d.plate, driverCar: d.car, driverRating: d.rating,
      createdAt: Date.now(),
    } : o));
    setDrivers((p) => p.map((x) => x.id === driverId ? { ...x, status: "busy" } : x));
  };

  const updateOrderStatus: Store["updateOrderStatus"] = (orderId, status) => {
    setOrders((p) => p.map((o) => o.id === orderId ? { ...o, status } : o));
    if (status === "completed" || status === "cancelled") {
      setOrders((prev) => {
        const o = prev.find((x) => x.id === orderId);
        if (o?.driverId) {
          setDrivers((d) => d.map((x) => x.id === o.driverId ? {
            ...x, status: "idle",
            todayEarnings: x.todayEarnings + (status === "completed" ? o.price : 0),
            totalOrders: x.totalOrders + (status === "completed" ? 1 : 0)
          } : x));
        }
        return prev;
      });
    }
  };

  const cancelOrder: Store["cancelOrder"] = (orderId) => updateOrderStatus(orderId, "cancelled");

  const setDriverOnline: Store["setDriverOnline"] = (driverId, online) => {
    setDrivers((p) => p.map((d) => d.id === driverId ? { ...d, online, status: online ? "idle" : "offline" } : d));
  };

  return (
    <StoreCtx.Provider value={{
      orders, drivers, complaints, currentDriverId,
      createOrder, acceptOrder, updateOrderStatus, cancelOrder, setDriverOnline,
    }}>
      {children}
    </StoreCtx.Provider>
  );
}

export function useStore() {
  const s = useContext(StoreCtx);
  if (!s) throw new Error("StoreProvider missing");
  return s;
}

export const statusLabel: Record<OrderStatus, string> = {
  pending: "等待接单",
  accepted: "司机前往中",
  arrived: "司机已到达",
  ongoing: "行程中",
  toPay: "待支付",
  completed: "已完成",
  cancelled: "已取消",
};

export const statusColor: Record<OrderStatus, string> = {
  pending: "bg-amber-100 text-amber-700",
  accepted: "bg-blue-100 text-blue-700",
  arrived: "bg-indigo-100 text-indigo-700",
  ongoing: "bg-violet-100 text-violet-700",
  toPay: "bg-pink-100 text-pink-700",
  completed: "bg-emerald-100 text-emerald-700",
  cancelled: "bg-gray-100 text-gray-600",
};