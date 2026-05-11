import { useState, useEffect, useRef } from "react";
import { PhoneFrame } from "./PhoneFrame";
import { AmapView } from "./AmapView";
import { useStore, Order, statusLabel, statusColor } from "../store";
import { ImageCropper } from "./ImageCropper";
import { uploadImage, dataURLtoFile, type BizType } from "../api/upload";
import { updateDriverProfile } from "../api/profile";
import { updateRealname } from "../api/realname";
import { updateLicense } from "../api/license";
import { updateVehicle } from "../api/vehicle";
import { getWallet, applyWithdraw, getBankCard, bindBankCard, updateBankCard, getWithdrawRecords, getIncomeDetail } from "../api/wallet";
import {
  Power, Home, ClipboardList, User, Star, TrendingUp, DollarSign,
  ChevronRight, Phone, Navigation, Shield, Award, Bell, Settings,
  MessageCircle, ChevronLeft, Zap, CheckCircle2, AlertTriangle, Car,
  Gift, Headphones, X, MapPin, Clock, CreditCard, AlertCircle,
  PhoneCall, FileText, Lock, LogOut, Activity, Volume2, WifiOff,
  Wifi, Copy, ChevronDown, Camera, Upload, IdCard, Truck
} from "lucide-react";

type Tab = "home" | "orders" | "me";
type MeStage = "main" | "income" | "service" | "car" | "sos" | "settings" | "realname" | "license" | "vehicle" | "wallet" | "bankcard" | "withdrawRecords";

/* ============================================================
   滑动确认组件
   ============================================================ */
function SlideToConfirm({
  label, onConfirm,
  gradient = "from-emerald-500 to-teal-500",
  emoji = "🚀", resetKey,
}: {
  label: string; onConfirm: () => void;
  gradient?: string; emoji?: string; resetKey?: string | number;
}) {
  const [offset, setOffset] = useState(0);
  const [confirmed, setConfirmed] = useState(false);
  const trackRef = useRef<HTMLDivElement>(null);
  const dragging = useRef(false);
  const startX = useRef(0);
  const startOffset = useRef(0);

  useEffect(() => { setOffset(0); setConfirmed(false); }, [resetKey]);

  const getMax = () => (trackRef.current ? trackRef.current.offsetWidth - 52 : 200);
  const onStart = (clientX: number) => {
    if (confirmed) return;
    dragging.current = true; startX.current = clientX; startOffset.current = offset;
  };
  const onMove = (clientX: number) => {
    if (!dragging.current) return;
    setOffset(Math.max(0, Math.min(clientX - startX.current + startOffset.current, getMax())));
  };
  const onEnd = () => {
    if (!dragging.current) return;
    dragging.current = false;
    if (offset >= getMax() * 0.82) { setOffset(getMax()); setConfirmed(true); setTimeout(onConfirm, 350); }
    else setOffset(0);
  };
  const pct = trackRef.current ? Math.min(100, (offset / getMax()) * 100) : 0;

  return (
    <div ref={trackRef} className="relative rounded-full overflow-hidden select-none"
      style={{ background: "#f0fdf4", height: 52 }}
      onMouseMove={e => onMove(e.clientX)} onMouseUp={onEnd} onMouseLeave={onEnd}
      onTouchMove={e => { e.preventDefault(); onMove(e.touches[0].clientX); }} onTouchEnd={onEnd}>
      <div className={`absolute left-0 top-0 h-full bg-gradient-to-r ${gradient} rounded-full`}
        style={{ width: `${6 + pct * 0.88}%`, opacity: 0.25, transition: dragging.current ? "none" : "width 0.3s ease" }} />
      <div className="absolute inset-0 flex items-center justify-center text-xs text-gray-400 pointer-events-none select-none">
        {confirmed
          ? <span className="text-emerald-600 font-medium">✓ 已确认</span>
          : <><span className="mr-14">{label}</span><span className="text-gray-300">›› 向右滑动</span></>}
      </div>
      <div className={`absolute top-1 left-1 w-11 h-11 bg-gradient-to-r ${gradient} rounded-full shadow-md flex items-center justify-center text-white text-base font-bold cursor-grab active:cursor-grabbing z-10`}
        style={{ transform: `translateX(${offset}px)`, transition: dragging.current ? "none" : "transform 0.3s cubic-bezier(.4,0,.2,1)" }}
        onMouseDown={e => { e.preventDefault(); onStart(e.clientX); }}
        onTouchStart={e => onStart(e.touches[0].clientX)}>
        {confirmed ? "✓" : emoji}
      </div>
    </div>
  );
}

/* ============================================================
   主 App
   ============================================================ */
export function DriverApp() {
  const { orders, drivers, currentDriverId, setDriverOnline, updateDriver, acceptOrder, updateOrderStatus } = useStore();
  const [tab, setTab] = useState<Tab>("home");
  const [meStage, setMeStage] = useState<MeStage>("main");
  const [toast, setToast] = useState<string | null>(null);
  const [selectedHistoryId, setSelectedHistoryId] = useState<string | null>(null);
  const [cancellationAlert, setCancellationAlert] = useState<{ name: string } | null>(null);
  const [withdrawalOpen, setWithdrawalOpen] = useState(false);
  const [sosOpen, setSosOpen] = useState(false);
  const [notifOpen, setNotifOpen] = useState(false);

  const driver = drivers.find(d => d.id === currentDriverId)!;
  const pendingOrders = orders.filter(o => o.status === "pending");
  const myActive = orders.find(o => o.driverId === currentDriverId && ["accepted", "arrived", "ongoing"].includes(o.status));
  const myHistory = orders.filter(o => o.driverId === currentDriverId && ["completed", "cancelled"].includes(o.status));

  // 检测乘客取消已接订单
  const prevActiveId = useRef<string | null>(null);
  useEffect(() => {
    if (prevActiveId.current && !myActive) {
      const prev = orders.find(o => o.id === prevActiveId.current);
      if (prev?.status === "cancelled") setCancellationAlert({ name: prev.passengerName });
    }
    prevActiveId.current = myActive?.id || null;
  }, [myActive?.id]);

  useEffect(() => {
    if (!toast) return;
    const t = setTimeout(() => setToast(null), 2000);
    return () => clearTimeout(t);
  }, [toast]);

  function showToast(msg: string) { setToast(msg); }
  function handleComplete() {
    if (myActive) { updateOrderStatus(myActive.id, "toPay"); setTimeout(() => myActive && updateOrderStatus(myActive.id, "completed"), 500); }
  }

  // 行程详情（全屏覆盖）
  if (selectedHistoryId) {
    const order = orders.find(o => o.id === selectedHistoryId);
    if (order) return (
      <PhoneFrame>
        <DriverTripDetailView order={order} onBack={() => setSelectedHistoryId(null)} onToast={showToast} />
      </PhoneFrame>
    );
  }

  return (
    <PhoneFrame>
      <div className="h-full flex flex-col bg-gray-50 relative">
        <div className="flex-1 overflow-y-auto">
          {tab === "home" && !myActive && (
            <DriverHome driver={driver} pendingOrders={pendingOrders}
              onToggle={() => setDriverOnline(driver.id, !driver.online)}
              onAccept={id => acceptOrder(id, driver.id)}
              onBell={() => setNotifOpen(true)} onToast={showToast} />
          )}
          {tab === "home" && myActive && (
            <DriverActive order={myActive}
              onAction={s => updateOrderStatus(myActive.id, s)}
              onComplete={handleComplete} onToast={showToast}
              onSOS={() => setSosOpen(true)} />
          )}
          {tab === "orders" && (
            <DriverOrders orders={myHistory}
              onOrderClick={id => setSelectedHistoryId(id)} onToast={showToast} />
          )}
          {tab === "me" && meStage === "main" && (
            <DriverMe driver={driver} onNav={s => setMeStage(s as MeStage)}
              onWithdraw={() => setWithdrawalOpen(true)} onToast={showToast}
              // 传入更新回调：头像上传成功后立即更新本地状态，无需重新加载 profile
              onUpdateDriver={(updates) => updateDriver(driver.id, updates)} />
          )}
          {tab === "me" && meStage === "income" && (
            <DriverIncomeView driver={driver} onBack={() => setMeStage("main")}
              onWithdraw={() => setWithdrawalOpen(true)} onToast={showToast} />
          )}
          {tab === "me" && meStage === "service" && (
            <DriverServiceView driver={driver} onBack={() => setMeStage("main")} onToast={showToast} />
          )}
          {tab === "me" && meStage === "car" && (
            <DriverCarView driver={driver} onBack={() => setMeStage("main")} onToast={showToast} />
          )}
          {tab === "me" && meStage === "sos" && (
            <DriverSOSView onBack={() => setMeStage("main")} onToast={showToast} />
          )}
          {tab === "me" && meStage === "settings" && (
            <DriverSettingsView driver={driver} onBack={() => setMeStage("main")}
              onToggleOnline={() => setDriverOnline(driver.id, !driver.online)} onToast={showToast} />
          )}
          {tab === "me" && meStage === "realname" && (
            <DriverRealnameView driver={driver} onBack={() => setMeStage("main")} onToast={showToast} />
          )}
          {tab === "me" && meStage === "license" && (
            <DriverLicenseView driver={driver} onBack={() => setMeStage("main")} onToast={showToast} />
          )}
          {tab === "me" && meStage === "vehicle" && (
            <DriverVehicleView driver={driver} onBack={() => setMeStage("main")} onToast={showToast} />
          )}
          {tab === "me" && meStage === "wallet" && (
            <DriverWalletView driver={driver} onBack={() => setMeStage("main")}
              onNav={(s: string) => setMeStage(s as MeStage)}
              onWithdraw={() => setWithdrawalOpen(true)} onToast={showToast} />
          )}
          {tab === "me" && meStage === "bankcard" && (
            <DriverBankCardView driver={driver} onBack={() => setMeStage("wallet")} onToast={showToast} />
          )}
          {tab === "me" && meStage === "withdrawRecords" && (
            <DriverWithdrawRecordsView onBack={() => setMeStage("wallet")} onToast={showToast} />
          )}
        </div>

        {/* 底部Tab */}
        <div className="flex border-t bg-white shrink-0">
          {[{ k: "home", icon: Home, label: "接单" }, { k: "orders", icon: ClipboardList, label: "订单" }, { k: "me", icon: User, label: "我的" }]
            .map(({ k, icon: Icon, label }) => (
              <button key={k} onClick={() => { setTab(k as Tab); if (k === "me") setMeStage("main"); }}
                className={`flex-1 py-2 flex flex-col items-center gap-0.5 relative ${tab === k ? "text-emerald-500" : "text-gray-400"}`}>
                <Icon className="w-5 h-5" />
                <span className="text-[10px]">{label}</span>
                {k === "orders" && tab !== "orders" && myHistory.filter(o => o.status === "completed" && Date.now() - o.createdAt < 3600000).length > 0 && (
                  <span className="absolute top-1 right-4 w-1.5 h-1.5 bg-emerald-500 rounded-full" />
                )}
              </button>
            ))}
        </div>

        {/* 新订单浮动提示（非首页时） */}
        {tab !== "home" && driver.online && pendingOrders.length > 0 && !cancellationAlert && (
          <FloatingNewOrderAlert order={pendingOrders[0]}
            onAccept={() => { acceptOrder(pendingOrders[0].id, driver.id); setTab("home"); }}
            onView={() => setTab("home")} />
        )}

        {/* 乘客取消弹窗 */}
        {cancellationAlert && (
          <CancellationModal name={cancellationAlert.name} onClose={() => setCancellationAlert(null)} />
        )}

        {/* SOS弹窗 */}
        {sosOpen && <DriverSOSModal onClose={() => setSosOpen(false)} onToast={showToast} />}

        {/* 提现弹窗 */}
        {withdrawalOpen && (
          <WithdrawalModal balance={driver.todayEarnings}
            onClose={() => setWithdrawalOpen(false)} onToast={showToast}
            onGoBankCard={() => { setWithdrawalOpen(false); setMeStage("bankcard"); }} />
        )}

        {/* 通知面板 */}
        {notifOpen && (
          <NotificationsPanel orders={myHistory} onClose={() => setNotifOpen(false)} />
        )}

        {/* Toast */}
        {toast && (
          <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-black/75 text-white text-xs px-4 py-2 rounded-lg z-[60]">
            {toast}
          </div>
        )}
      </div>
    </PhoneFrame>
  );
}

/* ============================================================
   接单首页
   ============================================================ */
function DriverHome({ driver, pendingOrders, onToggle, onAccept, onBell, onToast }: any) {
  const [pulse, setPulse] = useState(false);
  // Bottom sheet draggable panel state (expanded / collapsed)
  const MINI_PANEL_HEIGHT = 60;
  const EXPANDED_PANEL_HEIGHT = 420;
  const MAX_DRAG = EXPANDED_PANEL_HEIGHT - MINI_PANEL_HEIGHT;
  const [sheetState, setSheetState] = useState<"expanded" | "collapsed">("expanded");
  const [dragOffset, setDragOffset] = useState(0);
  const [isDragging, setIsDragging] = useState(false);
  const dragStartY = useRef<number>(0);
  const dragOffsetRef = useRef<number>(0);
  const draggingRef = useRef<boolean>(false);

  // Touch helpers
  const onPanelTouchStart = (e: any) => {
    draggingRef.current = true;
    setIsDragging(true);
    dragStartY.current = e.touches[0].clientY;
  };
  const onPanelTouchMove = (e: any) => {
    if (!draggingRef.current) return;
    const dy = e.touches[0].clientY - dragStartY.current;
    if (sheetState === "expanded") {
      const v = Math.max(0, Math.min(dy, MAX_DRAG));
      dragOffsetRef.current = v;
      setDragOffset(v);
    } else {
      const v = Math.min(0, dy);
      dragOffsetRef.current = v;
      setDragOffset(v);
    }
  };
  const onPanelTouchEnd = () => {
    const offset = dragOffsetRef.current;
    if (sheetState === "expanded") {
      if (offset >= 80) {
        setSheetState("collapsed");
        setDragOffset(0);
        dragOffsetRef.current = 0;
      } else {
        setDragOffset(0);
        dragOffsetRef.current = 0;
      }
    } else {
      if (offset <= -80) {
        setSheetState("expanded");
        setDragOffset(0);
        dragOffsetRef.current = 0;
      } else {
        setDragOffset(0);
        dragOffsetRef.current = 0;
      }
    }
    draggingRef.current = false;
    setIsDragging(false);
  };

  // Mouse helpers
  const onPanelMouseDown = (e: any) => {
    draggingRef.current = true;
    setIsDragging(true);
    dragStartY.current = e.clientY;
  };
  const onPanelMouseMove = (e: any) => {
    if (!draggingRef.current) return;
    const dy = e.clientY - dragStartY.current;
    if (sheetState === "expanded") {
      const v = Math.max(0, Math.min(dy, MAX_DRAG));
      dragOffsetRef.current = v;
      setDragOffset(v);
    } else {
      const v = Math.min(0, dy);
      dragOffsetRef.current = v;
      setDragOffset(v);
    }
  };
  const onPanelMouseUp = () => {
    const offset = dragOffsetRef.current;
    if (sheetState === "expanded") {
      if (offset >= 80) {
        setSheetState("collapsed");
        setDragOffset(0);
        dragOffsetRef.current = 0;
      } else {
        setDragOffset(0);
        dragOffsetRef.current = 0;
      }
    } else {
      if (offset <= -80) {
        setSheetState("expanded");
        setDragOffset(0);
        dragOffsetRef.current = 0;
      } else {
        setDragOffset(0);
        dragOffsetRef.current = 0;
      }
    }
    draggingRef.current = false;
    setIsDragging(false);
  };
  useEffect(() => {
    if (driver.online && pendingOrders.length > 0) { setPulse(true); const t = setTimeout(() => setPulse(false), 600); return () => clearTimeout(t); }
  }, [pendingOrders.length, driver.online]);

  return (
    <div className="relative h-full">
      <AmapView className="h-[280px]" showCar={driver.online} centerOnDriver />

      <div className="absolute top-2 left-3 right-3 bg-white rounded-2xl px-3 py-2.5 shadow-md flex items-center gap-3 z-10">
        <div className="w-10 h-10 rounded-full bg-emerald-100 flex items-center justify-center text-xl">👨‍✈️</div>
        <div className="flex-1">
          <div className="text-sm font-medium">{driver.name}</div>
          <div className="text-[10px] text-gray-500 flex items-center gap-1">
            <Star className="w-3 h-3 text-amber-400 fill-current" />{driver.rating}
            <span className="mx-1">·</span>{driver.plate}
          </div>
        </div>
        <button onClick={onToggle}
          className={`px-3.5 py-1.5 rounded-full text-xs flex items-center gap-1.5 transition-all ${driver.online ? "bg-emerald-500 text-white shadow-sm" : "bg-gray-200 text-gray-600"}`}>
          <Power className="w-3 h-3" />{driver.online ? "出车中" : "已收车"}
        </button>
      </div>

      <button onClick={onBell} className="absolute top-2 right-3 mt-[60px] w-9 h-9 bg-white rounded-full shadow flex items-center justify-center z-10">
        <Bell className="w-4 h-4 text-gray-600" />
        {pendingOrders.length > 0 && driver.online && <span className="absolute top-1 right-1 w-2 h-2 bg-rose-500 rounded-full" />}
      </button>

      <div className="absolute left-0 right-0 bottom-0 bg-white rounded-t-3xl shadow-2xl overflow-hidden" 
        style={{ height: sheetState === 'expanded' ? EXPANDED_PANEL_HEIGHT : MINI_PANEL_HEIGHT, transform: `translateY(${dragOffset}px)`, transition: isDragging ? "none" : "transform 0.3s ease-out, height 0.2s ease-out" }}
        onTouchStart={onPanelTouchStart} onTouchMove={onPanelTouchMove} onTouchEnd={onPanelTouchEnd}
        onMouseDown={onPanelMouseDown} onMouseMove={onPanelMouseMove} onMouseUp={onPanelMouseUp} onMouseLeave={onPanelMouseUp}
      >
        <div className="flex justify-center pt-2">
          <div className="w-9 h-1 bg-gray-400 rounded-full"></div>
        </div>
        {/* 今日收入卡 */}
        <div className="m-4 mb-3 bg-gradient-to-r from-emerald-500 to-teal-500 rounded-2xl p-4 text-white">
          <div className="flex items-center justify-between mb-2">
            <div className="text-xs opacity-90">今日收入</div>
            <button onClick={() => onToast("查看收入明细")} className="text-[10px] bg-white/20 px-2 py-0.5 rounded-full">明细 ›</button>
          </div>
          <div className="text-3xl font-light mb-2">¥{driver.todayEarnings.toFixed(2)}</div>
          <div className="grid grid-cols-3 gap-2 text-center text-[11px]">
            {[{ l: "📦 完单", v: driver.status === "busy" ? 9 : 8 }, { l: "⏱ 在线", v: "6.2h" }, { l: "⭐ 好评", v: "98%" }]
              .map(x => (
                <div key={x.l} className="bg-white/15 rounded-xl py-2">
                  <div className="opacity-90 mb-0.5">{x.l}</div>
                  <div className="font-medium">{x.v}</div>
                </div>
              ))}
          </div>
        </div>

        {/* 附近订单 */}
        <div className="px-4 flex items-center justify-between mb-2">
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium">附近订单</span>
            {driver.online && pendingOrders.length > 0 && (
              <span className={`text-[10px] px-2 py-0.5 rounded-full bg-rose-100 text-rose-600 ${pulse ? "scale-110" : ""} transition-transform`}>
                {pendingOrders.length} 单
              </span>
            )}
          </div>
          <span className="text-[10px] text-gray-400">自动刷新中</span>
        </div>

        <div className="space-y-2 max-h-[220px] overflow-y-auto px-4 pb-4">
          {!driver.online && (
            <div className="text-center py-8 text-xs text-gray-400">
              <div className="text-3xl mb-2">😴</div>
              <div>已收车，点击"已收车"重新出车</div>
            </div>
          )}
          {driver.online && pendingOrders.length === 0 && (
            <div className="text-center py-8 text-xs text-gray-400">
              <div className="text-3xl mb-2 animate-bounce">🚗</div>
              <div>正在为您智能派单...</div>
            </div>
          )}
          {driver.online && pendingOrders.map((o: Order) => (
            <div key={o.id} className="bg-orange-50 border border-orange-200 rounded-2xl p-3">
              <div className="flex justify-between items-center mb-2">
                <div className="flex items-center gap-2">
                  <span className="text-[10px] px-2 py-0.5 bg-orange-500 text-white rounded-full">{o.carType}</span>
                  {o.isPrebook && <span className="text-[10px] px-1.5 py-0.5 bg-purple-100 text-purple-600 rounded-full">预约</span>}
                  <span className="text-[10px] text-gray-500">距您 1.2km</span>
                </div>
                <span className="text-orange-600 font-medium">¥{o.price}</span>
              </div>
              <div className="flex gap-2 text-xs mb-0.5 items-start">
                <div className="w-2 h-2 rounded-full bg-emerald-500 mt-1 flex-shrink-0" />
                <span className="flex-1 text-gray-700 truncate">{o.from}</span>
              </div>
              <div className="flex gap-2 text-xs mb-1 items-start">
                <div className="w-2 h-2 rounded-full bg-rose-500 mt-1 flex-shrink-0" />
                <span className="flex-1 text-gray-700 truncate">{o.to}</span>
              </div>
              {o.note && (
                <div className="text-[10px] text-gray-500 bg-gray-50 rounded px-2 py-1 mb-1.5 truncate">备注：{o.note}</div>
              )}
              <div className="flex justify-between items-center mb-2">
                <span className="text-[10px] text-gray-500">{o.distanceKm}km · 约{o.estMinutes}分钟</span>
              </div>
              <SlideToConfirm label="立即抢单" emoji="🚗" gradient="from-orange-500 to-amber-500"
                onConfirm={() => onAccept(o.id)} resetKey={o.id} />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

/* ============================================================
   进行中订单
   ============================================================ */
function DriverActive({ order, onAction, onComplete, onToast, onSOS }: any) {
  const actionMap: any = {
    accepted: { label: "已到达上车点", next: "arrived", gradient: "from-indigo-500 to-blue-500", emoji: "📍" },
    arrived: { label: "开始行程", next: "ongoing", gradient: "from-violet-500 to-purple-500", emoji: "🚀" },
  };
  const action = actionMap[order.status];
  const statusInfo: any = {
    accepted: { title: "前往接驾中", sub: "请按导航路线行驶", color: "text-indigo-600", bg: "bg-indigo-50" },
    arrived: { title: "已到达上车点", sub: "等待乘客上车", color: "text-violet-600", bg: "bg-violet-50" },
    ongoing: { title: "行程进行中", sub: `目的地：${order.to}`, color: "text-emerald-600", bg: "bg-emerald-50" },
  };
  const info = statusInfo[order.status] || statusInfo.accepted;

  return (
    <div className="relative h-full">
      <AmapView className="h-[340px]" from={order.from} to={order.to} showCar centerOnDriver />

      {/* SOS浮动按钮 */}
      <button onClick={onSOS} className="absolute top-3 right-3 z-20 bg-rose-500 text-white text-xs font-bold px-3 py-1.5 rounded-full shadow-lg">
        SOS
      </button>

      <div className="absolute left-0 right-0 bottom-0 bg-white rounded-t-3xl p-4 shadow-2xl">
        <div className={`inline-flex items-center gap-1.5 text-xs px-3 py-1.5 rounded-full ${info.bg} ${info.color} mb-3`}>
          <Zap className="w-3 h-3" />{info.title} · {info.sub}
        </div>

        <div className="mb-3 space-y-1.5">
          <div className="flex gap-2 text-sm items-start">
            <div className="w-2 h-2 rounded-full bg-emerald-500 mt-1.5 flex-shrink-0" />
            <span className="flex-1 text-gray-800 truncate">{order.from}</span>
          </div>
          <div className="flex gap-2 text-sm items-start">
            <div className="w-2 h-2 rounded-full bg-rose-500 mt-1.5 flex-shrink-0" />
            <span className="flex-1 text-gray-800 truncate">{order.to}</span>
          </div>
        </div>

        {/* 乘客信息卡 */}
        <div className="bg-gray-50 rounded-2xl p-3 flex items-center gap-3 mb-3">
          <div className="w-10 h-10 rounded-full bg-white border border-gray-200 flex items-center justify-center text-xl">👤</div>
          <div className="flex-1">
            <div className="text-sm font-medium">{order.passengerName}</div>
            <div className="text-[10px] text-gray-500 mt-0.5">{order.passengerPhone}</div>
            {order.note && <div className="text-[10px] text-orange-500 mt-0.5">备注：{order.note}</div>}
          </div>
          <button onClick={() => onToast("正在联系乘客...")} className="w-9 h-9 bg-white border border-gray-200 rounded-full flex items-center justify-center">
            <MessageCircle className="w-4 h-4 text-gray-600" />
          </button>
          <button onClick={() => onToast("正在拨打...")} className="w-9 h-9 bg-emerald-500 rounded-full flex items-center justify-center shadow-sm">
            <Phone className="w-4 h-4 text-white" />
          </button>
        </div>

        {/* 工具栏 */}
        <div className="grid grid-cols-4 gap-2 mb-3 text-xs">
          {[
            { icon: Navigation, label: "导航", color: "text-blue-500", action: () => onToast("导航已启动") },
            { icon: Shield, label: "安全", color: "text-emerald-500", action: onSOS },
            { icon: Headphones, label: "客服", color: "text-violet-500", action: () => onToast("客服通道") },
            { icon: AlertTriangle, label: "异常", color: "text-amber-500", action: () => onToast("上报异常") },
          ].map(({ icon: Icon, label, color, action }) => (
            <button key={label} onClick={action} className="bg-gray-50 rounded-xl py-2.5 flex flex-col items-center gap-1 border border-gray-100">
              <Icon className={`w-4 h-4 ${color}`} /><span>{label}</span>
            </button>
          ))}
        </div>

        {action && (
          <SlideToConfirm label={action.label} emoji={action.emoji} gradient={action.gradient}
            onConfirm={() => onAction(action.next)} resetKey={order.status} />
        )}
        {order.status === "ongoing" && (
          <SlideToConfirm label={`完成行程 · ¥${order.price}`} emoji="🏁"
            gradient="from-orange-500 to-rose-500" onConfirm={onComplete}
            resetKey={order.id + "-ongoing"} />
        )}
      </div>
    </div>
  );
}

/* ============================================================
   订单历史
   ============================================================ */
function DriverOrders({ orders, onOrderClick, onToast }: { orders: Order[]; onOrderClick: (id: string) => void; onToast: (m: string) => void }) {
  const [tab, setTab] = useState<"today" | "all">("today");
  const completed = orders.filter(o => o.status === "completed");
  const today = completed.filter(o => Date.now() - o.createdAt < 86400000);
  const list = tab === "today" ? today : completed;
  const todayGMV = today.reduce((s, o) => s + o.price, 0);

  return (
    <div className="min-h-full bg-gray-50">
      <div className="bg-gradient-to-r from-emerald-500 to-teal-500 px-4 py-5">
        <div className="grid grid-cols-3 gap-3 text-white text-center">
          {[
            { val: today.length, label: "今日完单" },
            { val: `¥${todayGMV.toFixed(0)}`, label: "今日收入" },
            { val: completed.length, label: "历史总单" },
          ].map(c => (
            <div key={c.label} className="bg-white/20 rounded-xl py-3">
              <div className="text-xl font-medium">{c.val}</div>
              <div className="text-[10px] opacity-80 mt-0.5">{c.label}</div>
            </div>
          ))}
        </div>
      </div>

      <div className="flex bg-white border-b">
        {[["today", "今日"], ["all", "全部历史"]].map(([k, l]) => (
          <button key={k} onClick={() => setTab(k as any)}
            className={`flex-1 py-3 text-sm relative ${tab === k ? "text-emerald-600" : "text-gray-500"}`}>
            {l}
            {tab === k && <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-8 h-0.5 bg-emerald-500 rounded" />}
          </button>
        ))}
      </div>

      <div className="p-3 space-y-2">
        {list.map(o => (
          <button key={o.id} onClick={() => onOrderClick(o.id)}
            className="w-full bg-white rounded-2xl p-3.5 shadow-sm text-left active:bg-gray-50">
            <div className="flex justify-between mb-2">
              <span className="text-xs text-gray-500">{new Date(o.createdAt).toLocaleString("zh-CN")}</span>
              <div className="flex items-center gap-2">
                {o.rating && (
                  <span className="flex items-center gap-0.5 text-[10px] text-amber-500">
                    <Star className="w-3 h-3 fill-current" />{o.rating}分好评
                  </span>
                )}
                <span className="text-emerald-600 font-medium">+¥{o.price}</span>
              </div>
            </div>
            <div className="flex gap-2 text-xs mb-0.5 items-center">
              <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 flex-shrink-0" />
              <span className="flex-1 text-gray-700 truncate">{o.from}</span>
            </div>
            <div className="flex gap-2 text-xs items-center">
              <div className="w-1.5 h-1.5 rounded-full bg-rose-500 flex-shrink-0" />
              <span className="flex-1 text-gray-700 truncate">{o.to}</span>
            </div>
            <div className="flex justify-between items-center mt-2.5 pt-2.5 border-t border-gray-50 text-[10px] text-gray-400">
              <span>{o.carType} · {o.distanceKm}km · {o.estMinutes}分钟</span>
              <ChevronRight className="w-3.5 h-3.5 text-gray-300" />
            </div>
          </button>
        ))}
        {list.length === 0 && (
          <div className="text-center py-12 text-gray-400 text-sm">
            <div className="text-3xl mb-2">📋</div>暂无订单记录
          </div>
        )}
      </div>
    </div>
  );
}

/* ============================================================
   行程详情（司机视角）
   ============================================================ */
function DriverTripDetailView({ order, onBack, onToast }: { order: Order; onBack: () => void; onToast: (m: string) => void }) {
  const platform = (order.price * 0.2).toFixed(2);
  const net = (order.price * 0.8).toFixed(2);
  const timeline = [
    { time: new Date(order.createdAt).toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" }), label: "乘客下单", done: true },
    { time: "已接单", label: "接受订单", done: !!order.driverId },
    { time: "已到达", label: "到达上车点", done: ["arrived", "ongoing", "completed"].includes(order.status) },
    { time: "行程中", label: "开始行程", done: ["ongoing", "completed"].includes(order.status) },
    { time: "已完成", label: "完成行程", done: order.status === "completed" },
  ];

  return (
    <div className="min-h-full bg-gray-50">
      {/* 顶栏 */}
      <div className="bg-white px-4 py-3 flex items-center border-b sticky top-0 z-10">
        <button onClick={onBack}><ChevronLeft className="w-5 h-5" /></button>
        <div className="flex-1 text-center font-medium">行程详情</div>
        <button onClick={() => onToast("已复制订单号")} className="text-xs text-gray-400">
          <Copy className="w-4 h-4" />
        </button>
      </div>

      {/* 地图 */}
      <AmapView className="h-40" from={order.from} to={order.to} />

      {/* 状态 + 收入 */}
      <div className="mx-3 -mt-4 relative z-10 bg-gradient-to-r from-emerald-500 to-teal-500 rounded-2xl p-4 text-white shadow-lg mb-3">
        <div className="flex items-center justify-between mb-2">
          <span className={`text-[10px] px-2 py-0.5 rounded-full bg-white/20`}>{statusLabel[order.status]}</span>
          <span className="text-xs opacity-80">{new Date(order.createdAt).toLocaleString("zh-CN")}</span>
        </div>
        <div className="text-3xl font-light mb-1">+¥{order.price}</div>
        <div className="text-xs opacity-80">订单号：{order.id.slice(-8)}</div>
      </div>

      {/* 路线 */}
      <div className="mx-3 bg-white rounded-2xl p-4 shadow-sm mb-3">
        <div className="text-xs font-medium text-gray-500 mb-3">行程路线</div>
        <div className="space-y-3">
          <div className="flex gap-3 items-start">
            <div className="w-2 h-2 rounded-full bg-emerald-500 mt-1 flex-shrink-0" />
            <div><div className="text-sm text-gray-800">{order.from}</div><div className="text-[10px] text-gray-400">上车点</div></div>
          </div>
          <div className="flex gap-3 items-start">
            <div className="w-2 h-2 rounded-full bg-rose-500 mt-1 flex-shrink-0" />
            <div><div className="text-sm text-gray-800">{order.to}</div><div className="text-[10px] text-gray-400">下车点</div></div>
          </div>
        </div>
        <div className="flex gap-4 mt-3 pt-3 border-t border-gray-50 text-xs text-gray-500">
          <span>📏 {order.distanceKm}km</span>
          <span>⏱ {order.estMinutes}分钟</span>
          <span>🚗 {order.carType}</span>
        </div>
      </div>

      {/* 乘客信息 */}
      <div className="mx-3 bg-white rounded-2xl p-4 shadow-sm mb-3">
        <div className="text-xs font-medium text-gray-500 mb-3">乘客信息</div>
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-full bg-violet-100 flex items-center justify-center text-xl">👤</div>
          <div className="flex-1">
            <div className="text-sm font-medium">{order.passengerName}</div>
            <div className="text-[10px] text-gray-400">{order.passengerPhone}</div>
          </div>
          {order.rating && (
            <div className="text-right">
              <div className="flex items-center gap-0.5 justify-end">
                {[1, 2, 3, 4, 5].map(i => (
                  <Star key={i} className={`w-3 h-3 ${i <= order.rating! ? "text-amber-400 fill-current" : "text-gray-200"}`} />
                ))}
              </div>
              {order.ratingComment && <div className="text-[10px] text-gray-500 mt-0.5">"{order.ratingComment}"</div>}
              {order.ratingTags?.map((t: string) => (
                <span key={t} className="text-[9px] bg-amber-50 text-amber-600 px-1.5 py-0.5 rounded mr-1">{t}</span>
              ))}
            </div>
          )}
        </div>
        {order.note && <div className="mt-2 text-[10px] text-orange-500 bg-orange-50 rounded-lg px-2 py-1">备注：{order.note}</div>}
      </div>

      {/* 收入明细 */}
      <div className="mx-3 bg-white rounded-2xl p-4 shadow-sm mb-3">
        <div className="text-xs font-medium text-gray-500 mb-3">收入明细</div>
        {[
          { label: "行程费用", val: `¥${order.price}`, color: "text-gray-800" },
          { label: "平台服务费 (20%)", val: `-¥${platform}`, color: "text-rose-500" },
          { label: "实际到账", val: `¥${net}`, color: "text-emerald-600" },
        ].map(item => (
          <div key={item.label} className="flex justify-between py-2 border-b last:border-0">
            <span className="text-sm text-gray-600">{item.label}</span>
            <span className={`text-sm font-medium ${item.color}`}>{item.val}</span>
          </div>
        ))}
        {order.paymentMethod && (
          <div className="flex justify-between py-2 text-xs text-gray-400">
            <span>支付方式</span><span>{order.paymentMethod}</span>
          </div>
        )}
      </div>

      {/* 行程时间线 */}
      <div className="mx-3 bg-white rounded-2xl p-4 shadow-sm mb-3">
        <div className="text-xs font-medium text-gray-500 mb-3">行程轨迹</div>
        <div className="space-y-3">
          {timeline.map((t, i) => (
            <div key={i} className="flex items-center gap-3">
              <div className={`w-5 h-5 rounded-full flex items-center justify-center flex-shrink-0 ${t.done ? "bg-emerald-500" : "bg-gray-100"}`}>
                {t.done ? <CheckCircle2 className="w-3 h-3 text-white" /> : <div className="w-2 h-2 rounded-full bg-gray-300" />}
              </div>
              <div className="flex-1">
                <span className={`text-sm ${t.done ? "text-gray-800" : "text-gray-400"}`}>{t.label}</span>
              </div>
              <span className="text-[10px] text-gray-400">{t.time}</span>
            </div>
          ))}
        </div>
      </div>

      {/* 操作按钮 */}
      <div className="mx-3 mb-6 grid grid-cols-2 gap-3">
        <button onClick={() => onToast("开具电子发票")} className="bg-white rounded-2xl py-3 text-sm text-blue-500 border border-blue-100 shadow-sm flex items-center justify-center gap-2">
          <FileText className="w-4 h-4" />开具发票
        </button>
        <button onClick={() => onToast("已发起投诉")} className="bg-white rounded-2xl py-3 text-sm text-gray-500 border border-gray-100 shadow-sm flex items-center justify-center gap-2">
          <AlertTriangle className="w-4 h-4" />问题反馈
        </button>
      </div>
    </div>
  );
}

/* ============================================================
   司机个人中心
   ============================================================ */
// 司机个人中心页（头像、统计卡片、功能入口列表）
// onUpdateDriver: 头像/昵称等字段变更后回调，用于即时更新本地状态
function DriverMe({ driver, onNav, onWithdraw, onToast, onUpdateDriver }: any) {
  const [cropperOpen, setCropperOpen] = useState(false);
  const [cropperImage, setCropperImage] = useState<string | null>(null);
  const [uploading, setUploading] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleAvatarClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = () => {
      setCropperImage(reader.result as string);
      setCropperOpen(true);
    };
    reader.readAsDataURL(file);
    e.target.value = "";
  };

  // 头像裁剪确认：上传到 MinIO → 调用 UpdateProfile API 写 DB → 更新本地状态
  const handleCropConfirm = async (croppedDataURL: string) => {
    setCropperOpen(false);
    setUploading(true);
    try {
      // 1. 将裁剪后的 DataURL 转 File，上传至 MinIO
      const file = dataURLtoFile(croppedDataURL, "avatar.jpg");
      const result = await uploadImage(file, "avatar" as BizType);
      if (result?.url) {
        // 2. 将图片 URL 写入后端 drivers.avatar 字段
        const res = await updateDriverProfile({
          driver_id: 200000001,
          avatar: result.url,
        });
        if (res?.success) {
          // 3. 写入成功后立即更新本地 React 状态，头像即时显示无需刷新
          onUpdateDriver?.({ avatar: result.url });
          onToast("头像更新成功");
        } else {
          onToast(res?.message || "头像绑定失败");
        }
      } else {
        onToast("头像上传失败");
      }
    } catch {
      onToast("头像上传异常");
    }
    setUploading(false);
  };

  return (
    <div>
      <div className="bg-gradient-to-b from-emerald-500 to-teal-600 px-4 pt-6 pb-8 relative overflow-hidden">
        <div className="absolute -right-4 -top-4 w-32 h-32 rounded-full bg-white/10" />
        <div className="absolute right-8 bottom-0 w-20 h-20 rounded-full bg-white/5" />
        <div className="flex items-center gap-3 mb-5 relative">
          <button onClick={handleAvatarClick} className="relative group">
            {driver.avatar ? (
              <img src={driver.avatar} alt="头像" className="w-16 h-16 rounded-full border-2 border-white/50 object-cover" />
            ) : (
              <div className="w-16 h-16 rounded-full bg-white/30 border-2 border-white/50 flex items-center justify-center text-3xl">👨‍✈️</div>
            )}
            <div className="absolute inset-0 rounded-full bg-black/30 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
              <Camera className="w-5 h-5 text-white" />
            </div>
            {uploading && (
              <div className="absolute inset-0 rounded-full bg-black/50 flex items-center justify-center">
                <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
              </div>
            )}
          </button>
          <input ref={fileInputRef} type="file" accept="image/jpeg,image/png" className="hidden" onChange={handleFileSelect} />
          <div className="flex-1">
            <div className="text-white font-medium text-lg flex items-center gap-2">
              {driver.name}
              <span className="bg-amber-400 text-gray-800 text-[10px] px-1.5 py-0.5 rounded font-bold">金牌司机</span>
            </div>
            <div className="text-white/80 text-xs mt-1">{driver.plate} · {driver.car}</div>
          </div>
          <button onClick={() => onNav("settings")} className="w-9 h-9 rounded-full bg-white/20 flex items-center justify-center">
            <Settings className="w-4 h-4 text-white" />
          </button>
        </div>
        <div className="grid grid-cols-3 gap-2 relative">
          {[{ val: driver.rating, label: "综合评分" }, { val: driver.totalOrders, label: "总完单" }, { val: "98%", label: "好评率" }]
            .map(item => (
              <div key={item.label} className="bg-white/20 rounded-xl py-3 text-center text-white">
                <div className="text-lg font-medium">{item.val}</div>
                <div className="text-[10px] opacity-80 mt-0.5">{item.label}</div>
              </div>
            ))}
        </div>
      </div>

      {/* 收入卡 */}
      <div className="mx-3 -mt-4 relative z-10 bg-white rounded-2xl shadow-md p-4 flex items-center justify-between mb-3">
        <div>
          <div className="text-xs text-gray-500">今日可结收入</div>
          <div className="text-emerald-600 font-medium text-xl mt-0.5">¥{driver.todayEarnings.toFixed(2)}</div>
        </div>
        <div className="flex gap-2">
          <button onClick={() => onNav("income")} className="bg-gray-100 text-gray-600 text-xs px-3 py-2 rounded-full">查看明细</button>
          <button onClick={onWithdraw} className="bg-emerald-500 text-white text-xs px-3 py-2 rounded-full">立即提现</button>
        </div>
      </div>

      <div className="px-3 pb-6">
        <div className="bg-white rounded-2xl overflow-hidden">
          {[
            { icon: DollarSign, label: "收入明细", val: `¥${(driver.todayEarnings * 12).toFixed(0)}`, key: "income", color: "text-emerald-500" },
            { icon: CreditCard, label: "我的钱包", val: "", key: "wallet", color: "text-teal-500" },
            { icon: TrendingUp, label: "服务分", val: `${Math.floor(driver.rating * 20)}/100`, key: "service", color: "text-blue-500" },
            { icon: Award, label: "司机等级", val: "金牌 Lv.3", key: null, color: "text-amber-500" },
            { icon: IdCard, label: "实名认证", val: "已认证", key: "realname", color: "text-cyan-500" },
            { icon: FileText, label: "驾驶证认证", val: "已认证", key: "license", color: "text-indigo-500" },
            { icon: Car, label: "我的车辆", val: driver.car.split("·")[0].trim(), key: "vehicle", color: "text-violet-500" },
            { icon: Shield, label: "安全中心", val: "", key: "sos", color: "text-teal-500" },
            { icon: Gift, label: "司机福利", val: "", key: null, color: "text-rose-500" },
            { icon: Settings, label: "设置", val: "", key: "settings", color: "text-gray-500" },
            { icon: Headphones, label: "联系客服", val: "", key: null, color: "text-indigo-500" },
          ].map(({ icon: Icon, label, val, key, color }) => (
            <button key={label} onClick={() => key ? onNav(key) : onToast(label)}
              className="w-full flex items-center gap-3 px-4 py-3.5 border-b border-gray-50 last:border-0 hover:bg-gray-50">
              <div className={`w-9 h-9 rounded-full bg-gray-50 flex items-center justify-center ${color}`}>
                <Icon className="w-4 h-4" />
              </div>
              <span className="flex-1 text-left text-sm text-gray-900">{label}</span>
              {val && <span className="text-xs text-gray-400">{val}</span>}
              <ChevronRight className="w-4 h-4 text-gray-300" />
            </button>
          ))}
        </div>
      </div>

      {/* 头像裁剪弹窗 */}
      {cropperOpen && cropperImage && (
        <ImageCropper
          imageSrc={cropperImage}
          aspectRatio="square"
          onConfirm={handleCropConfirm}
          onCancel={() => setCropperOpen(false)}
        />
      )}
    </div>
  );
}

/* ============================================================
   收入明细页
   ============================================================ */
const TEST_DRIVER_ID = 200000001;

function DriverIncomeView({ driver, onBack, onWithdraw, onToast }: any) {
  const [period, setPeriod] = useState<"today" | "week" | "month">("today");
  const [income, setIncome] = useState<number>(driver.todayEarnings);
  const [trips, setTrips] = useState<number>(8);
  const [hours, setHours] = useState<number>(6.2);
  const [trend, setTrend] = useState<{ date: string; income: number }[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    async function loadIncome() {
      setLoading(true);
      const data = await fetch(
        `http://localhost:8080/api/v1/driver/income?driver_id=${TEST_DRIVER_ID}&period=${period}`
      ).then((res) => res.json());
      if (data?.summary) {
        setIncome(data.summary.income || 0);
        setTrips(data.summary.order_count || 0);
        setHours((data.summary.online_duration || 0) / 3600);
      }
      if (data?.trend) {
        setTrend(data.trend);
      }
      setLoading(false);
    }
    loadIncome();
  }, [period]);

  const hourlyIncome = hours > 0 ? income / hours : 0;
  const base = income * 0.85;
  const bonus = income * 0.1;
  const subsidy = income * 0.05;

  // 计算趋势图最大高度（仅 week/month 需要）
  const maxIncome = Math.max(...trend.map((t) => t.income), 1);

  // 根据周期获取图表标题
  const chartTitle = period === "week" ? "本周收入趋势" : "本月收入趋势";

  return (
    <div className="min-h-full bg-gray-50">
      <div className="bg-white px-4 py-3 flex items-center border-b">
        <button onClick={onBack}><ChevronLeft className="w-5 h-5" /></button>
        <div className="flex-1 text-center font-medium">收入明细</div>
        <button onClick={onWithdraw} className="text-emerald-600 text-sm">提现</button>
      </div>

      <div className="bg-gradient-to-r from-emerald-500 to-teal-500 p-5 text-white text-center">
        <div className="text-xs opacity-90 mb-1">
          {period === "today" ? "今日收入 (元)" : `${period === "week" ? "本周" : "本月"}总收入 (元)`}
        </div>
        <div className="text-4xl font-light mb-3">¥{income.toFixed(2)}</div>
        <div className="flex justify-center gap-2 mb-4">
          {[["today", "今日"], ["week", "本周"], ["month", "本月"]].map(([k, l]) => (
            <button key={k} onClick={() => setPeriod(k as any)}
              className={`px-4 py-1.5 rounded-full text-xs ${period === k ? "bg-white text-emerald-600 font-medium" : "bg-white/20"}`}>{l}</button>
          ))}
        </div>
        <div className="grid grid-cols-3 gap-2 text-center text-[11px]">
          {[{ v: `${trips}单`, l: "完单数" }, { v: `${hours.toFixed(1)}h`, l: "出车时长" }, { v: `¥${hourlyIncome.toFixed(1)}`, l: "时均收入" }]
            .map(x => (
              <div key={x.l} className="bg-white/15 rounded-xl py-2">
                <div className="font-medium">{x.v}</div>
                <div className="opacity-80 mt-0.5">{x.l}</div>
              </div>
            ))}
        </div>
      </div>

      {/* ========== 今日视图：大数字卡片（无柱形图）========== */}
      {period === "today" && (
        <div className="mx-3 mt-3 bg-white rounded-2xl p-6 shadow-sm">
          <div className="text-sm font-medium text-gray-400 mb-4">今日收入详情</div>
          <div className="text-center">
            <div className="text-5xl font-bold text-emerald-600 mb-2">¥{income.toFixed(2)}</div>
            <div className="text-xs text-gray-400 mt-4">
              {new Date().toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric', weekday: 'long' })}
            </div>
          </div>
          {/* 今日快捷指标 */}
          <div className="grid grid-cols-3 gap-3 mt-6 pt-4 border-t border-gray-100">
            <div className="text-center">
              <div className="text-lg font-semibold text-gray-800">{trips}</div>
              <div className="text-[10px] text-gray-400 mt-0.5">完成订单</div>
            </div>
            <div className="text-center border-x border-gray-100">
              <div className="text-lg font-semibold text-gray-800">{hours.toFixed(1)}</div>
              <div className="text-[10px] text-gray-400 mt-0.5">在线时长(h)</div>
            </div>
            <div className="text-center">
              <div className="text-lg font-semibold text-emerald-600">¥{hourlyIncome.toFixed(1)}</div>
              <div className="text-[10px] text-gray-400 mt-0.5">时均收入</div>
            </div>
          </div>
        </div>
      )}

      {/* ========== 本周 / 本月视图：动态柱形图 ========== */}
      {(period === "week" || period === "month") && (
        <div className="mx-3 mt-3 bg-white rounded-2xl p-4 shadow-sm">
          <div className="flex items-center justify-between mb-3">
            <div className="text-sm font-medium">{chartTitle}</div>
            {trend.length > 0 && (
              <div className="text-xs text-gray-400">最高 ¥{maxIncome.toFixed(0)}</div>
            )}
          </div>
          {/* 柱形图区域 - 本月数据多时支持横向滚动 */}
          <div className={`relative ${period === "month" ? "h-40" : "h-36"} pt-6 pb-1`}>
            {/* 背景网格线 */}
            {[0, 25, 50, 75].map(pct => (
              <div key={pct} className="absolute left-0 right-0 border-b border-dashed border-gray-100"
                   style={{ top: `${100 - pct}%`, transform: 'translateY(50%)' }} />
            ))}
            <div className={`flex items-end ${trend.length > 7 ? "gap-1" : "gap-2"} h-full px-1 overflow-x-auto`}>
              {trend.length > 0 ? trend.map((t, i) => {
                const barH = Math.max(6, (t.income / maxIncome) * 100);
                const dayLabel = (() => {
                  const d = new Date(t.date);
                  if (period === "week") {
                    const weekdays = ["日", "一", "二", "三", "四", "五", "六"];
                    return `周${weekdays[d.getDay()]}`;
                  }
                  return `${d.getDate()}`;
                })();
                const isToday = i === trend.length - 1;
                const hasData = t.income > 0;
                return (
                  <div key={i} className={`flex-shrink-0 flex flex-col items-center gap-1 relative group
                                       ${period === "month" ? "w-7 md:w-9" : "flex-1 max-w-[48px]"}`}>
                    {/* 金额标签 - 今日加大加粗 */}
                    {hasData && (
                      <div className={`absolute left-1/2 -translate-x-1/2 whitespace-nowrap px-1.5 py-0.5 rounded-md transition-all group-hover:scale-110
                                      ${isToday
                                        ? 'text-xs font-bold text-white bg-emerald-500 shadow-md shadow-emerald-300/50 -top-1'
                                        : 'text-[9px] font-medium text-emerald-600 bg-emerald-50 top-0'}`}>
                        ¥{t.income.toFixed(0)}
                      </div>
                    )}
                    {/* 柱体 - 今日突出 + 往期淡化 */}
                    <div
                      className={`w-full rounded-lg transition-all duration-300 relative
                                ${isToday
                                  ? 'bg-gradient-to-t from-emerald-600 to-emerald-400 shadow-lg shadow-emerald-400/40 ring-2 ring-emerald-300'
                                  : hasData
                                    ? 'bg-gradient-to-t from-emerald-300/60 to-emerald-100/40'
                                    : 'bg-gray-50'}`}
                      style={{
                        height: `${isToday ? Math.max(barH, 12) : barH}%`,
                        minHeight: hasData ? (isToday ? '20px' : '8px') : '4px',
                        marginTop: hasData ? (isToday ? '22px' : '18px') : 'auto',
                      }}
                    >
                      {/* 光泽层 */}
                      {hasData && (
                        <div className={`absolute inset-x-0 top-0 h-1/3 rounded-t-lg bg-gradient-to-b from-white/40 to-transparent ${isToday ? 'opacity-60' : 'opacity-30'}`} />
                      )}
                      {/* 今日底部发光 */}
                      {isToday && (
                        <div className="absolute -bottom-1 left-1/2 -translate-x-1/2 w-4/5 h-1.5 bg-emerald-300/50 rounded-full blur-sm" />
                      )}
                    </div>
                    {/* 日期标签 - 今日带"今天"标记 */}
                    <div className="flex flex-col items-center">
                      <div className={`text-[9px] whitespace-nowrap ${isToday ? 'font-bold text-emerald-600' : 'text-gray-400'}`}>
                        {dayLabel}
                      </div>
                      {isToday && (
                        <span className="text-[8px] font-medium text-emerald-500 scale-90">今天</span>
                      )}
                    </div>
                  </div>
                );
              }) : (
                /* 无数据时的骨架屏 */
                Array.from({ length: period === "week" ? 7 : Math.min(30, new Date().getDate()) }).map((_, i) => (
                  <div key={i} className={`flex-shrink-0 flex flex-col items-center gap-1
                                        ${period === "month" ? "w-6" : "flex-1"}`}>
                    <div className="w-full max-w-[24px] rounded-md bg-gray-100 animate-pulse"
                         style={{ height: `${Math.random() * 60 + 15}%`, marginTop: '16px' }} />
                    <div className="text-[9px] text-gray-300">{period === "week" ? ["一","二","三","四","五","六","日"][i] : i + 1}</div>
                  </div>
                ))
              )}
            </div>
          </div>
          {/* 图例 + 统计 */}
          <div className="flex items-center justify-between mt-2 pt-2 border-t border-gray-50">
            <div className="flex items-center gap-3">
              <div className="flex items-center gap-1 text-[10px] text-gray-400">
                <span className="w-2 h-2 rounded bg-gradient-to-t from-emerald-400/70 to-emerald-200/60 inline-block" /> 有收入
              </div>
              <div className="flex items-center gap-1 text-[10px] text-gray-400">
                <span className="w-2 h-2 rounded bg-gray-100 inline-block" /> 无收入
              </div>
            </div>
            <div className="text-[10px] text-gray-300">
              共 {trend.length} 天 · 日均 ¥{(income / Math.max(trend.length, 1)).toFixed(0)}
            </div>
          </div>
        </div>
      )}

      {/* 收入构成 */}
      <div className="mx-3 mt-3 bg-white rounded-2xl p-4 shadow-sm">
        <div className="text-sm font-medium mb-3">收入构成</div>
        {[
          { name: "行程基础收入", val: base, color: "bg-emerald-500", pct: 85 },
          { name: "奖励收入", val: bonus, color: "bg-amber-500", pct: 10 },
          { name: "活动补贴", val: subsidy, color: "bg-blue-500", pct: 5 },
        ].map(item => (
          <div key={item.name} className="py-2.5 border-b last:border-0">
            <div className="flex justify-between mb-1.5">
              <div className="flex items-center gap-2">
                <div className={`w-2 h-2 rounded-full ${item.color}`} />
                <span className="text-sm text-gray-700">{item.name}</span>
              </div>
              <span className="text-sm font-medium text-gray-900">¥{item.val.toFixed(2)}</span>
            </div>
            <div className="h-1.5 bg-gray-100 rounded-full overflow-hidden ml-4">
              <div className={`h-full ${item.color} rounded-full`} style={{ width: `${item.pct}%` }} />
            </div>
          </div>
        ))}
      </div>

      {/* 结算记录 */}
      <div className="mx-3 mt-3 mb-6 bg-white rounded-2xl p-4 shadow-sm">
        <div className="text-sm font-medium mb-3">最近结算记录</div>
        {[
          { date: "2026-04-22", amount: 289.5, status: "已结算" },
          { date: "2026-04-21", amount: 312.0, status: "已结算" },
          { date: "2026-04-20", amount: 245.8, status: "已结算" },
        ].map((r, i) => (
          <div key={i} className="flex justify-between items-center py-2.5 border-b last:border-0">
            <div className="text-sm text-gray-800">{r.date}</div>
            <div className="flex items-center gap-3">
              <span className="text-[10px] bg-emerald-100 text-emerald-700 px-2 py-0.5 rounded">{r.status}</span>
              <span className="text-emerald-600 font-medium">¥{r.amount.toFixed(2)}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

/* ============================================================
   服务分页
   ============================================================ */
function DriverServiceView({ driver, onBack, onToast }: any) {
  const score = Math.floor(driver.rating * 20);
  const items = [
    { name: "行程评分", score: 98, desc: "基于乘客评价综合计算" },
    { name: "接单率", score: 94, desc: "分配订单中成功接单比例" },
    { name: "完单率", score: 99, desc: "接单后成功完成比例" },
    { name: "安全驾驶", score: 96, desc: "基于行程轨迹分析" },
  ];

  return (
    <div className="min-h-full bg-gray-50">
      <div className="bg-white px-4 py-3 flex items-center border-b">
        <button onClick={onBack}><ChevronLeft className="w-5 h-5" /></button>
        <div className="flex-1 text-center font-medium">服务分</div>
        <div className="w-5" />
      </div>

      <div className="bg-gradient-to-br from-blue-500 to-indigo-600 p-6 text-white text-center">
        <div className="text-xs opacity-80 mb-2">综合服务分</div>
        <div className="text-6xl font-bold mb-1">{score}</div>
        <div className="text-sm opacity-90">{score >= 90 ? "🏆 优秀" : score >= 80 ? "👍 良好" : "📊 待提升"}</div>
        <div className="mt-4 text-xs opacity-80">服务分影响您的派单优先级和奖励金额</div>
      </div>

      <div className="mx-3 mt-3 bg-white rounded-2xl p-4 shadow-sm">
        <div className="text-sm font-medium mb-3">分项明细</div>
        {items.map(item => (
          <div key={item.name} className="py-3 border-b last:border-0">
            <div className="flex justify-between items-center mb-1.5">
              <span className="text-sm text-gray-800">{item.name}</span>
              <span className="text-sm font-medium text-blue-600">{item.score}<span className="text-gray-400 text-xs">/100</span></span>
            </div>
            <div className="h-1.5 bg-gray-100 rounded-full overflow-hidden mb-1">
              <div className="h-full bg-gradient-to-r from-blue-400 to-indigo-500 rounded-full" style={{ width: `${item.score}%` }} />
            </div>
            <div className="text-[10px] text-gray-400">{item.desc}</div>
          </div>
        ))}
      </div>

      <div className="mx-3 mt-3 mb-6 bg-amber-50 border border-amber-200 rounded-2xl p-4">
        <div className="text-sm font-medium text-amber-700 mb-2 flex items-center gap-1.5">
          <Zap className="w-4 h-4" /> 提升建议
        </div>
        <div className="space-y-2 text-xs text-amber-600">
          {["提高接单率：保持良好在线率，及时响应派单", "主动服务：主动与乘客沟通，提升体验", "安全驾驶：避免急刹急加速，保持平稳行驶"]
            .map(t => (
              <div key={t} className="flex items-start gap-2">
                <CheckCircle2 className="w-3.5 h-3.5 mt-0.5 flex-shrink-0" />{t}
              </div>
            ))}
        </div>
      </div>
    </div>
  );
}

/* ============================================================
   车辆信息页
   ============================================================ */
function DriverCarView({ driver, onBack, onToast }: any) {
  return (
    <div className="min-h-full bg-gray-50">
      <div className="bg-white px-4 py-3 flex items-center border-b">
        <button onClick={onBack}><ChevronLeft className="w-5 h-5" /></button>
        <div className="flex-1 text-center font-medium">我的车辆</div>
        <button onClick={() => onToast("申请修改")} className="text-blue-500 text-sm">修改</button>
      </div>

      <div className="mx-3 mt-4 bg-gradient-to-br from-gray-800 to-gray-900 rounded-2xl p-5 text-white relative overflow-hidden">
        <div className="text-3xl mb-1">🚗</div>
        <div className="text-xl font-bold">{driver.plate}</div>
        <div className="text-sm opacity-80 mt-1">{driver.car}</div>
        <div className="absolute -right-3 -bottom-3 text-8xl opacity-10">🚗</div>
      </div>

      <div className="mx-3 mt-3 bg-white rounded-2xl overflow-hidden shadow-sm">
        {[
          { label: "车牌号", val: driver.plate },
          { label: "车辆品牌", val: driver.car.split("·")[0].trim() },
          { label: "车身颜色", val: driver.car.split("·")[1]?.trim() || "银色" },
          { label: "座位数", val: "5座" },
          { label: "排放标准", val: "国六" },
          { label: "营运证号", val: "运营J2026042" },
          { label: "保险到期", val: "2027-03-31" },
          { label: "年检到期", val: "2027-06-30" },
        ].map((item, i, arr) => (
          <div key={item.label} className={`flex items-center px-4 py-3.5 ${i < arr.length - 1 ? "border-b border-gray-50" : ""}`}>
            <span className="flex-1 text-sm text-gray-500">{item.label}</span>
            <span className="text-sm text-gray-900">{item.val}</span>
          </div>
        ))}
      </div>

      <div className="mx-3 mt-3 mb-6 grid grid-cols-2 gap-3">
        <button onClick={() => onToast("查看年检记录")} className="bg-white rounded-2xl py-3.5 text-sm text-blue-500 border border-blue-100 shadow-sm">年检记录 ›</button>
        <button onClick={() => onToast("保险续保")} className="bg-white rounded-2xl py-3.5 text-sm text-emerald-500 border border-emerald-100 shadow-sm">保险续保 ›</button>
      </div>
    </div>
  );
}

/* ============================================================
   安全中心
   ============================================================ */
function DriverSOSView({ onBack, onToast }: { onBack: () => void; onToast: (m: string) => void }) {
  const [recording, setRecording] = useState(false);
  return (
    <div className="min-h-full bg-gray-50">
      <div className="bg-white px-4 py-3 flex items-center border-b">
        <button onClick={onBack}><ChevronLeft className="w-5 h-5" /></button>
        <div className="flex-1 text-center font-medium">安全中心</div>
        <div className="w-5" />
      </div>

      {/* SOS紧急按钮 */}
      <div className="bg-gradient-to-br from-rose-500 to-red-600 mx-3 mt-4 rounded-2xl p-5 text-white text-center">
        <button onClick={() => onToast("已发出SOS求助信号！正在联系紧急联系人...")}
          className="w-20 h-20 rounded-full bg-white/20 border-4 border-white/50 flex items-center justify-center mx-auto mb-3 active:scale-95 transition-transform">
          <span className="text-white font-bold text-2xl">SOS</span>
        </button>
        <div className="text-sm font-medium mb-1">紧急求助</div>
        <div className="text-[11px] opacity-80">遇到危险时按下，将通知紧急联系人和平台</div>
      </div>

      {/* 快捷联系 */}
      <div className="mx-3 mt-3 bg-white rounded-2xl overflow-hidden shadow-sm">
        <div className="px-4 py-3 border-b">
          <div className="text-sm font-medium">紧急联系</div>
        </div>
        {[
          { icon: "🚔", label: "报警 110", sub: "公安机关紧急求助", action: () => onToast("正在拨打110...") },
          { icon: "🚑", label: "急救 120", sub: "医疗紧急救援", action: () => onToast("正在拨打120...") },
          { icon: "🎧", label: "平台紧急客服", sub: "7×24小时专属通道", action: () => onToast("正在接入客服...") },
        ].map(item => (
          <button key={item.label} onClick={item.action}
            className="w-full flex items-center gap-3 px-4 py-3.5 border-b last:border-0 hover:bg-gray-50">
            <div className="w-10 h-10 rounded-full bg-rose-50 flex items-center justify-center text-xl">{item.icon}</div>
            <div className="flex-1 text-left">
              <div className="text-sm font-medium text-gray-900">{item.label}</div>
              <div className="text-[10px] text-gray-400 mt-0.5">{item.sub}</div>
            </div>
            <PhoneCall className="w-4 h-4 text-gray-300" />
          </button>
        ))}
      </div>

      {/* 行车记录 */}
      <div className="mx-3 mt-3 bg-white rounded-2xl p-4 shadow-sm">
        <div className="flex items-center justify-between mb-3">
          <div className="text-sm font-medium">行驶记录仪</div>
          <button onClick={() => { setRecording(!recording); onToast(recording ? "已停止录制" : "录制已开启"); }}
            className={`flex items-center gap-1.5 text-xs px-3 py-1.5 rounded-full ${recording ? "bg-rose-100 text-rose-600" : "bg-gray-100 text-gray-600"}`}>
            <Volume2 className="w-3 h-3" />{recording ? "录制中" : "开启录制"}
          </button>
        </div>
        <div className="text-[11px] text-gray-500 bg-gray-50 rounded-xl px-3 py-2">
          行程录制功能保护您和乘客的权益，录音文件将加密存储72小时
        </div>
      </div>

      {/* 安全知识 */}
      <div className="mx-3 mt-3 mb-6 bg-blue-50 border border-blue-100 rounded-2xl p-4">
        <div className="text-sm font-medium text-blue-700 mb-2">安全驾驶提示</div>
        {["行程中请勿使用手机", "不得绕路或不走最优路线", "出现纠纷请第一时间联系平台", "遇乘客异常行为保持冷静，及时上报"].map(t => (
          <div key={t} className="flex items-start gap-2 text-xs text-blue-600 mb-1.5">
            <CheckCircle2 className="w-3.5 h-3.5 mt-0.5 flex-shrink-0 text-blue-500" />{t}
          </div>
        ))}
      </div>
    </div>
  );
}

/* ============================================================
   设置页
   ============================================================ */
function DriverSettingsView({ driver, onBack, onToggleOnline, onToast }: any) {
  const [sound, setSound] = useState(true);
  const [autoNav, setAutoNav] = useState(true);
  const [nightMode, setNightMode] = useState(false);
  const [changeMobileOpen, setChangeMobileOpen] = useState(false);
  const [changePasswordOpen, setChangePasswordOpen] = useState(false);

  return (
    <div className="min-h-full bg-gray-50">
      <div className="bg-white px-4 py-3 flex items-center border-b">
        <button onClick={onBack}><ChevronLeft className="w-5 h-5" /></button>
        <div className="flex-1 text-center font-medium">设置</div>
        <div className="w-5" />
      </div>

      {/* 在线状态 */}
      <div className="mx-3 mt-4 bg-white rounded-2xl overflow-hidden shadow-sm">
        <div className="px-4 py-3 border-b text-xs text-gray-400 font-medium">出车设置</div>
        <div className="flex items-center px-4 py-3.5 border-b">
          <div className="flex-1">
            <div className="text-sm text-gray-900">出车状态</div>
            <div className="text-[10px] text-gray-400 mt-0.5">{driver.online ? "当前出车中，可接收订单" : "已收车"}</div>
          </div>
          <button onClick={onToggleOnline}
            className={`relative w-12 h-6 rounded-full transition-colors ${driver.online ? "bg-emerald-500" : "bg-gray-200"}`}>
            <div className={`absolute top-0.5 w-5 h-5 rounded-full bg-white shadow transition-transform ${driver.online ? "translate-x-6" : "translate-x-0.5"}`} />
          </button>
        </div>
      </div>

      {/* 通知设置 */}
      <div className="mx-3 mt-3 bg-white rounded-2xl overflow-hidden shadow-sm">
        <div className="px-4 py-3 border-b text-xs text-gray-400 font-medium">通知与声音</div>
        {[
          { label: "新订单提示音", sub: "收到新订单时播放提示", val: sound, set: setSound },
          { label: "自动导航", sub: "接单后自动启动导航", val: autoNav, set: setAutoNav },
          { label: "夜间模式", sub: "23:00-06:00自动降低亮度", val: nightMode, set: setNightMode },
        ].map(item => (
          <div key={item.label} className="flex items-center px-4 py-3.5 border-b last:border-0">
            <div className="flex-1">
              <div className="text-sm text-gray-900">{item.label}</div>
              <div className="text-[10px] text-gray-400 mt-0.5">{item.sub}</div>
            </div>
            <button onClick={() => item.set(!item.val)}
              className={`relative w-12 h-6 rounded-full transition-colors ${item.val ? "bg-emerald-500" : "bg-gray-200"}`}>
              <div className={`absolute top-0.5 w-5 h-5 rounded-full bg-white shadow transition-transform ${item.val ? "translate-x-6" : "translate-x-0.5"}`} />
            </button>
          </div>
        ))}
      </div>

      {/* 账户 */}
      <div className="mx-3 mt-3 bg-white rounded-2xl overflow-hidden shadow-sm">
        <div className="px-4 py-3 border-b text-xs text-gray-400 font-medium">账户</div>
        {[
          { icon: Phone, label: "修改手机号", action: () => setChangeMobileOpen(true) },
          { icon: Lock, label: "修改密码", action: () => setChangePasswordOpen(true) },
          { icon: Shield, label: "隐私设置", action: () => onToast("隐私设置") },
          { icon: FileText, label: "用户协议", action: () => onToast("用户协议") },
        ].map(({ icon: Icon, label, action }) => (
          <button key={label} onClick={action} className="w-full flex items-center px-4 py-3.5 border-b last:border-0 hover:bg-gray-50">
            <Icon className="w-4 h-4 text-gray-400 mr-3" />
            <span className="flex-1 text-left text-sm text-gray-900">{label}</span>
            <ChevronRight className="w-4 h-4 text-gray-300" />
          </button>
        ))}
      </div>

      <div className="mx-3 mt-3 mb-6">
        <button onClick={() => onToast("已退出登录")} className="w-full bg-white rounded-2xl py-3.5 text-sm text-rose-500 border border-rose-100 shadow-sm flex items-center justify-center gap-2">
          <LogOut className="w-4 h-4" />退出登录
        </button>
      </div>

      {/* 修改手机号弹窗 */}
      {changeMobileOpen && (
        <ChangeMobileModal
          onClose={() => setChangeMobileOpen(false)}
          onToast={onToast}
        />
      )}

      {/* 修改密码弹窗 */}
      {changePasswordOpen && (
        <ChangePasswordModal
          onClose={() => setChangePasswordOpen(false)}
          onToast={onToast}
        />
      )}
    </div>
  );
}

/* ============================================================
   浮动新订单提醒
   ============================================================ */
function FloatingNewOrderAlert({ order, onAccept, onView }: { order: Order; onAccept: () => void; onView: () => void }) {
  return (
    <div className="absolute bottom-16 left-3 right-3 bg-white rounded-2xl shadow-2xl border border-orange-200 z-40 overflow-hidden">
      <div className="bg-gradient-to-r from-orange-500 to-amber-500 px-3 py-1.5 flex items-center justify-between">
        <div className="flex items-center gap-1.5 text-white text-xs font-medium">
          <Activity className="w-3.5 h-3.5" />新订单抢单中
        </div>
        <span className="text-white text-xs opacity-80">{order.carType}</span>
      </div>
      <div className="p-3">
        <div className="flex justify-between items-center mb-2">
          <div className="flex gap-2 text-xs flex-col flex-1 mr-2">
            <div className="flex gap-1.5 items-center">
              <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 flex-shrink-0" />
              <span className="text-gray-700 truncate">{order.from}</span>
            </div>
            <div className="flex gap-1.5 items-center">
              <div className="w-1.5 h-1.5 rounded-full bg-rose-500 flex-shrink-0" />
              <span className="text-gray-700 truncate">{order.to}</span>
            </div>
          </div>
          <div className="text-orange-600 font-bold text-lg">¥{order.price}</div>
        </div>
        <SlideToConfirm label="抢单" emoji="🚗" gradient="from-orange-500 to-amber-500"
          onConfirm={onAccept} resetKey={order.id + "-float"} />
        <button onClick={onView} className="w-full text-center text-xs text-gray-400 mt-2">查看详情 ›</button>
      </div>
    </div>
  );
}

/* ============================================================
   乘客取消弹窗
   ============================================================ */
function CancellationModal({ name, onClose }: { name: string; onClose: () => void }) {
  return (
    <div className="absolute inset-0 bg-black/50 z-50 flex items-center justify-center p-6">
      <div className="bg-white rounded-2xl p-5 w-full shadow-2xl">
        <div className="text-center mb-4">
          <div className="text-4xl mb-2">😔</div>
          <div className="font-medium text-gray-900">乘客已取消订单</div>
          <div className="text-sm text-gray-500 mt-1">{name} 已取消了本次行程</div>
        </div>
        <div className="bg-amber-50 rounded-xl p-3 text-xs text-amber-700 mb-4">
          取消费将在24小时内原路返还，若乘客有违规取消行为，平台将给予相应补偿。
        </div>
        <button onClick={onClose} className="w-full bg-gradient-to-r from-emerald-500 to-teal-500 text-white py-3 rounded-full font-medium">
          知道了，继续接单
        </button>
      </div>
    </div>
  );
}

/* ============================================================
   SOS 弹窗（行程中）
   ============================================================ */
function DriverSOSModal({ onClose, onToast }: { onClose: () => void; onToast: (m: string) => void }) {
  return (
    <div className="absolute inset-0 bg-black/60 z-50 flex items-end">
      <div className="bg-white rounded-t-3xl w-full p-5">
        <div className="flex items-center justify-between mb-4">
          <div className="font-medium text-gray-900 text-lg">紧急求助</div>
          <button onClick={onClose}><X className="w-5 h-5 text-gray-500" /></button>
        </div>
        <div className="grid grid-cols-3 gap-3 mb-4">
          {[
            { icon: "🚔", label: "报警 110", color: "bg-blue-50 border-blue-200", action: () => { onToast("正在拨打110..."); onClose(); } },
            { icon: "🚑", label: "急救 120", color: "bg-rose-50 border-rose-200", action: () => { onToast("正在拨打120..."); onClose(); } },
            { icon: "🎧", label: "紧急客服", color: "bg-violet-50 border-violet-200", action: () => { onToast("接入紧急客服..."); onClose(); } },
          ].map(item => (
            <button key={item.label} onClick={item.action}
              className={`${item.color} border rounded-2xl py-4 flex flex-col items-center gap-2`}>
              <span className="text-2xl">{item.icon}</span>
              <span className="text-xs text-gray-700">{item.label}</span>
            </button>
          ))}
        </div>
        <button onClick={() => { onToast("已向平台发送SOS！"); onClose(); }}
          className="w-full bg-rose-500 text-white py-3.5 rounded-full font-medium text-base">
          🆘 向平台发送SOS求助
        </button>
        <div className="text-[10px] text-center text-gray-400 mt-2">平台将立即响应并联系紧急联系人</div>
      </div>
    </div>
  );
}

/* ============================================================
   提现弹窗
   ============================================================ */
function WithdrawalModal({ balance, onClose, onToast, onGoBankCard }: { balance: number; onClose: () => void; onToast: (m: string) => void; onGoBankCard?: () => void }) {
  const [amount, setAmount] = useState(String(balance.toFixed(2)));
  const [card, setCard] = useState("加载中...");
  const [hasCard, setHasCard] = useState(false);
  const [walletData, setWalletData] = useState<any>(null);

  useEffect(() => {
    async function load() {
      const w = await getWallet(200000001);
      if (w) {
        setWalletData(w);
        setCard(w.has_bank_card ? `${w.bank_card_no}` : "未绑定银行卡");
        setHasCard(w.has_bank_card);
      } else {
        setCard("未绑定银行卡");
      }
    }
    load();
  }, []);

  const realBalance = walletData?.balance ?? balance;

  function handleConfirm() {
    const v = parseFloat(amount);
    if (!v || v <= 0) { onToast("请输入有效金额"); return; }
    if (v > realBalance) { onToast("提现金额不能超过可用余额"); return; }
    if (v > 5000) { onToast("单笔提现上限5000元"); return; }
    if (!hasCard) { onToast("请先绑定银行卡"); return; }
    if (walletData && walletData.today_withdraw_count >= 3) { onToast("今日提现次数已达上限(3次)"); return; }
    applyWithdraw(200000001, v).then(res => {
      if (res?.success) {
        onToast(`¥${v.toFixed(2)} 提现申请已提交，预计2小时到账`);
        onClose();
      } else {
        onToast(res?.message || "提现失败");
      }
    });
  }

  return (
    <div className="absolute inset-0 bg-black/50 z-50 flex items-end">
      <div className="bg-white rounded-t-3xl w-full p-5">
        <div className="flex items-center justify-between mb-4">
          <div className="font-medium text-gray-900">提现</div>
          <button onClick={onClose}><X className="w-5 h-5 text-gray-500" /></button>
        </div>

        <div className="bg-emerald-50 rounded-2xl p-3 mb-4 text-center">
          <div className="text-xs text-gray-500">可提现余额</div>
          <div className="text-3xl font-light text-emerald-600 mt-1">¥{realBalance.toFixed(2)}</div>
          {walletData?.frozen_amount > 0 && (
            <div className="text-[10px] text-amber-600 mt-1">冻结金额：¥{walletData.frozen_amount.toFixed(2)}（T-3结算中）</div>
          )}
        </div>

        <div className="mb-3">
          <div className="text-xs text-gray-500 mb-1.5">提现金额</div>
          <div className="flex items-center border border-gray-200 rounded-xl px-3 py-2.5 focus-within:border-emerald-500">
            <span className="text-gray-400 mr-1.5">¥</span>
            <input type="number" value={amount} onChange={e => setAmount(e.target.value)}
              className="flex-1 outline-none text-lg font-medium text-gray-900" />
            <button onClick={() => setAmount(String(realBalance.toFixed(2)))} className="text-xs text-emerald-500">全部</button>
          </div>
        </div>

        <button onClick={onGoBankCard} className="flex items-center gap-2 bg-gray-50 rounded-xl px-3 py-2.5 mb-4 w-full">
          <CreditCard className="w-4 h-4 text-gray-400" />
          <span className={`flex-1 text-sm text-left ${hasCard ? "text-gray-700" : "text-gray-400"}`}>{card}</span>
          <ChevronDown className="w-4 h-4 text-gray-300" />
        </button>

        <div className="text-[10px] text-gray-400 mb-2">提现规则：每天最多3次 · 单笔上限¥5000 · 2小时内到账</div>
        {walletData && (
          <div className="text-[10px] text-amber-500 mb-4">今日已提现 {walletData.today_withdraw_count}/3 次</div>
        )}

        <button onClick={handleConfirm}
          className="w-full bg-gradient-to-r from-emerald-500 to-teal-500 text-white py-3.5 rounded-full font-medium">
          确认提现 ¥{parseFloat(amount) > 0 ? parseFloat(amount).toFixed(2) : "0.00"}
        </button>
      </div>
    </div>
  );
}

/* ============================================================
   通知面板
   ============================================================ */
function NotificationsPanel({ orders, onClose }: { orders: Order[]; onClose: () => void }) {
  const recents = orders.slice(0, 5);
  const systemNotifs = [
    { icon: "🎁", title: "本周完单奖励", body: "您本周完成18单，获得额外奖励¥50！", time: "10分钟前" },
    { icon: "⚡", title: "早高峰调度", body: "明日07:00-09:00早高峰补贴提高至1.5倍", time: "1小时前" },
    { icon: "📋", title: "服务分更新", body: "您的服务分已更新为98分，继续保持！", time: "3小时前" },
  ];

  return (
    <div className="absolute inset-0 bg-black/50 z-50 flex">
      <div className="flex-1" onClick={onClose} />
      <div className="w-[85%] bg-white h-full flex flex-col">
        <div className="px-4 py-3 border-b flex items-center justify-between">
          <div className="font-medium">消息通知</div>
          <button onClick={onClose}><X className="w-5 h-5 text-gray-500" /></button>
        </div>
        <div className="flex-1 overflow-y-auto">
          <div className="px-4 py-2 text-xs text-gray-400 font-medium mt-2">系统消息</div>
          {systemNotifs.map((n, i) => (
            <div key={i} className="flex gap-3 px-4 py-3 border-b hover:bg-gray-50">
              <div className="w-10 h-10 rounded-full bg-gray-100 flex items-center justify-center text-xl flex-shrink-0">{n.icon}</div>
              <div className="flex-1 min-w-0">
                <div className="text-sm font-medium text-gray-900">{n.title}</div>
                <div className="text-xs text-gray-500 mt-0.5">{n.body}</div>
                <div className="text-[10px] text-gray-400 mt-1">{n.time}</div>
              </div>
            </div>
          ))}
          <div className="px-4 py-2 text-xs text-gray-400 font-medium mt-2">近期完单</div>
          {recents.length === 0 && <div className="text-center text-gray-400 text-sm py-6">暂无记录</div>}
          {recents.map(o => (
            <div key={o.id} className="flex gap-3 px-4 py-3 border-b">
              <div className="w-10 h-10 rounded-full bg-emerald-100 flex items-center justify-center text-xl flex-shrink-0">✅</div>
              <div className="flex-1 min-w-0">
                <div className="text-sm text-gray-800 truncate">{o.from} → {o.to}</div>
                <div className="text-xs text-emerald-600 mt-0.5">+¥{o.price}</div>
                <div className="text-[10px] text-gray-400 mt-0.5">{new Date(o.createdAt).toLocaleString("zh-CN")}</div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

/* ============================================================
   修改手机号弹窗
   ============================================================ */
function ChangeMobileModal({ onClose, onToast }: { onClose: () => void; onToast: (m: string) => void }) {
  const [newMobile, setNewMobile] = useState("");
  const [verifyCode, setVerifyCode] = useState("");
  const [countdown, setCountdown] = useState(0);

  useEffect(() => {
    if (countdown > 0) {
      const t = setTimeout(() => setCountdown(countdown - 1), 1000);
      return () => clearTimeout(t);
    }
  }, [countdown]);

  const sendCode = () => {
    if (!newMobile || newMobile.length !== 11) {
      onToast("请输入正确的手机号");
      return;
    }
    onToast("验证码已发送");
    setCountdown(60);
  };

  const handleSubmit = async () => {
    if (!newMobile || !verifyCode) {
      onToast("请填写完整信息");
      return;
    }
    const res = await fetch("http://localhost:8080/api/v1/driver/mobile", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ driver_id: 200000001, new_mobile: newMobile, verify_code: verifyCode }),
    }).then(r => r.json());
    if (res?.success) {
      onToast("手机号修改成功");
      onClose();
    } else {
      onToast(res?.message || "修改失败");
    }
  };

  return (
    <div className="absolute inset-0 bg-black/50 z-50 flex items-end">
      <div className="bg-white rounded-t-3xl w-full p-5">
        <div className="flex items-center justify-between mb-4">
          <div className="font-medium text-gray-900">修改手机号</div>
          <button onClick={onClose}><X className="w-5 h-5 text-gray-500" /></button>
        </div>
        <div className="text-xs text-amber-600 bg-amber-50 rounded-lg p-2 mb-4">1个月内最多修改1次手机号</div>
        <div className="mb-3">
          <div className="text-xs text-gray-500 mb-1.5">新手机号</div>
          <input type="tel" value={newMobile} onChange={e => setNewMobile(e.target.value)}
            placeholder="请输入新手机号"
            className="w-full border border-gray-200 rounded-xl px-3 py-2.5 text-sm outline-none focus:border-emerald-500" />
        </div>
        <div className="mb-4">
          <div className="text-xs text-gray-500 mb-1.5">验证码</div>
          <div className="flex gap-2">
            <input type="text" value={verifyCode} onChange={e => setVerifyCode(e.target.value)}
              placeholder="请输入验证码"
              className="flex-1 border border-gray-200 rounded-xl px-3 py-2.5 text-sm outline-none focus:border-emerald-500" />
            <button onClick={sendCode} disabled={countdown > 0}
              className={`px-4 py-2.5 rounded-xl text-xs ${countdown > 0 ? "bg-gray-100 text-gray-400" : "bg-emerald-500 text-white"}`}>
              {countdown > 0 ? `${countdown}s` : "获取验证码"}
            </button>
          </div>
        </div>
        <button onClick={handleSubmit} className="w-full bg-gradient-to-r from-emerald-500 to-teal-500 text-white py-3 rounded-full font-medium">确认修改</button>
      </div>
    </div>
  );
}

/* ============================================================
   修改密码弹窗
   ============================================================ */
function ChangePasswordModal({ onClose, onToast }: { onClose: () => void; onToast: (m: string) => void }) {
  const [oldPassword, setOldPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");

  const handleSubmit = async () => {
    if (!oldPassword || !newPassword || !confirmPassword) {
      onToast("请填写完整信息");
      return;
    }
    if (newPassword !== confirmPassword) {
      onToast("两次密码输入不一致");
      return;
    }
    if (newPassword === oldPassword) {
      onToast("新密码不能与原密码相同");
      return;
    }
    const res = await fetch("http://localhost:8080/api/v1/driver/password", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ driver_id: 200000001, old_password: oldPassword, new_password: newPassword }),
    }).then(r => r.json());
    if (res?.success) {
      onToast("密码修改成功");
      onClose();
    } else {
      onToast(res?.message || "修改失败");
    }
  };

  return (
    <div className="absolute inset-0 bg-black/50 z-50 flex items-end">
      <div className="bg-white rounded-t-3xl w-full p-5">
        <div className="flex items-center justify-between mb-4">
          <div className="font-medium text-gray-900">修改密码</div>
          <button onClick={onClose}><X className="w-5 h-5 text-gray-500" /></button>
        </div>
        <div className="text-xs text-amber-600 bg-amber-50 rounded-lg p-2 mb-4">1小时内最多修改5次密码</div>
        <div className="mb-3">
          <div className="text-xs text-gray-500 mb-1.5">原密码</div>
          <input type="password" value={oldPassword} onChange={e => setOldPassword(e.target.value)}
            placeholder="请输入原密码"
            className="w-full border border-gray-200 rounded-xl px-3 py-2.5 text-sm outline-none focus:border-emerald-500" />
        </div>
        <div className="mb-3">
          <div className="text-xs text-gray-500 mb-1.5">新密码</div>
          <input type="password" value={newPassword} onChange={e => setNewPassword(e.target.value)}
            placeholder="请输入新密码"
            className="w-full border border-gray-200 rounded-xl px-3 py-2.5 text-sm outline-none focus:border-emerald-500" />
        </div>
        <div className="mb-4">
          <div className="text-xs text-gray-500 mb-1.5">确认新密码</div>
          <input type="password" value={confirmPassword} onChange={e => setConfirmPassword(e.target.value)}
            placeholder="请再次输入新密码"
            className="w-full border border-gray-200 rounded-xl px-3 py-2.5 text-sm outline-none focus:border-emerald-500" />
        </div>
        <button onClick={handleSubmit} className="w-full bg-gradient-to-r from-emerald-500 to-teal-500 text-white py-3 rounded-full font-medium">确认修改</button>
      </div>
    </div>
  );
}

/* ============================================================
   通用照片上传组件（用于认证照片）
   ============================================================ */
function PhotoUploader({ label, bizType, currentUrl, onUploaded, onToast }: {
  label: string; bizType: BizType; currentUrl?: string;
  onUploaded: (url: string) => void; onToast: (m: string) => void;
}) {
  const [cropperOpen, setCropperOpen] = useState(false);
  const [cropperImage, setCropperImage] = useState<string | null>(null);
  const [uploading, setUploading] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleClick = () => fileInputRef.current?.click();

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = () => { setCropperImage(reader.result as string); setCropperOpen(true); };
    reader.readAsDataURL(file);
    e.target.value = "";
  };

  const handleCropConfirm = async (croppedDataURL: string) => {
    setCropperOpen(false);
    setUploading(true);
    try {
      const file = dataURLtoFile(croppedDataURL, `${bizType}.jpg`);
      const result = await uploadImage(file, bizType);
      if (result?.url) {
        onUploaded(result.url);
        onToast("上传成功");
      } else {
        onToast("上传失败");
      }
    } catch { onToast("上传异常"); }
    setUploading(false);
  };

  return (
    <div>
      <div className="text-xs text-gray-500 mb-1.5">{label}</div>
      <input ref={fileInputRef} type="file" accept="image/jpeg,image/png" className="hidden" onChange={handleFileSelect} />
      <button onClick={handleClick} className="w-full aspect-[16/10] rounded-xl border-2 border-dashed border-gray-200 flex flex-col items-center justify-center gap-2 overflow-hidden relative hover:border-emerald-300 transition-colors">
        {currentUrl ? (
          <img src={currentUrl} alt={label} className="w-full h-full object-cover" />
        ) : (
          <>
            {uploading ? (
              <div className="w-8 h-8 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin" />
            ) : (
              <>
                <Upload className="w-6 h-6 text-gray-300" />
                <span className="text-xs text-gray-400">点击上传</span>
              </>
            )}
          </>
        )}
      </button>
      {cropperOpen && cropperImage && (
        <ImageCropper
          imageSrc={cropperImage}
          aspectRatio="portrait"
          onConfirm={handleCropConfirm}
          onCancel={() => setCropperOpen(false)}
        />
      )}
    </div>
  );
}

/* ============================================================
   实名认证页
   ============================================================ */
function DriverRealnameView({ driver, onBack, onToast }: any) {
  const [realName, setRealName] = useState("");
  const [idCardNo, setIdCardNo] = useState("");
  const [idCardFrontUrl, setIdCardFrontUrl] = useState("");
  const [idCardBackUrl, setIdCardBackUrl] = useState("");
  const [submitting, setSubmitting] = useState(false);

  // 从后端已提交的数据回填
  useEffect(() => {
    if (driver.realnameInfo) {
      if (driver.realnameInfo.real_name) setRealName(driver.realnameInfo.real_name);
      if (driver.realnameInfo.id_card_front_url) setIdCardFrontUrl(driver.realnameInfo.id_card_front_url);
      if (driver.realnameInfo.id_card_back_url) setIdCardBackUrl(driver.realnameInfo.id_card_back_url);
    }
  }, []);

  const handleSubmit = async () => {
    if (!realName || !idCardNo || !idCardFrontUrl || !idCardBackUrl) {
      onToast("请填写完整信息并上传身份证正反面");
      return;
    }
    setSubmitting(true);
    const res = await updateRealname({
      driver_id: 200000001,
      real_name: realName,
      id_card_no: idCardNo,
      id_card_front_url: idCardFrontUrl,
      id_card_back_url: idCardBackUrl,
    });
    setSubmitting(false);
    if (res?.success) {
      onToast("实名认证提交成功，等待审核");
      onBack();
    } else {
      onToast(res?.message || "提交失败");
    }
  };

  return (
    <div className="min-h-full bg-gray-50">
      <div className="bg-white px-4 py-3 flex items-center border-b">
        <button onClick={onBack}><ChevronLeft className="w-5 h-5" /></button>
        <div className="flex-1 text-center font-medium">实名认证</div>
        <div className="w-5" />
      </div>

      <div className="p-4 space-y-4">
        <div className="bg-blue-50 border border-blue-100 rounded-xl p-3 text-xs text-blue-600">
          请确保上传的身份证照片清晰完整，审核通常1-3个工作日完成
        </div>

        <div>
          <div className="text-xs text-gray-500 mb-1.5">真实姓名</div>
          <input type="text" value={realName} onChange={e => setRealName(e.target.value)}
            placeholder="请输入身份证上的姓名"
            className="w-full border border-gray-200 rounded-xl px-3 py-2.5 text-sm outline-none focus:border-emerald-500" />
        </div>

        <div>
          <div className="text-xs text-gray-500 mb-1.5">身份证号</div>
          <input type="text" value={idCardNo} onChange={e => setIdCardNo(e.target.value)}
            placeholder="请输入18位身份证号" maxLength={18}
            className="w-full border border-gray-200 rounded-xl px-3 py-2.5 text-sm outline-none focus:border-emerald-500" />
        </div>

        <PhotoUploader label="身份证正面（人像面）" bizType="idcard" currentUrl={idCardFrontUrl}
          onUploaded={setIdCardFrontUrl} onToast={onToast} />

        <PhotoUploader label="身份证反面（国徽面）" bizType="idcard" currentUrl={idCardBackUrl}
          onUploaded={setIdCardBackUrl} onToast={onToast} />

        <button onClick={handleSubmit} disabled={submitting}
          className="w-full bg-gradient-to-r from-emerald-500 to-teal-500 text-white py-3 rounded-full font-medium disabled:opacity-50">
          {submitting ? "提交中..." : "提交认证"}
        </button>
      </div>
    </div>
  );
}

/* ============================================================
   驾驶证认证页
   ============================================================ */
function DriverLicenseView({ driver, onBack, onToast }: any) {
  const [licenseNo, setLicenseNo] = useState("");
  const [licenseType, setLicenseType] = useState("C1");
  const [licenseUrl, setLicenseUrl] = useState("");
  const [submitting, setSubmitting] = useState(false);

  // 从后端已提交的数据回填
  useEffect(() => {
    if (driver.licenseInfo) {
      if (driver.licenseInfo.license_no) setLicenseNo(driver.licenseInfo.license_no);
      if (driver.licenseInfo.license_type) setLicenseType(driver.licenseInfo.license_type);
      if (driver.licenseInfo.license_url) setLicenseUrl(driver.licenseInfo.license_url);
    }
  }, []);

  const handleSubmit = async () => {
    if (!licenseNo || !licenseUrl) {
      onToast("请填写驾驶证号并上传驾驶证照片");
      return;
    }
    setSubmitting(true);
    const res = await updateLicense({
      driver_id: 200000001,
      license_no: licenseNo,
      license_type: licenseType,
      license_url: licenseUrl,
    });
    setSubmitting(false);
    if (res?.success) {
      onToast("驾驶证认证提交成功，等待审核");
      onBack();
    } else {
      onToast(res?.message || "提交失败");
    }
  };

  return (
    <div className="min-h-full bg-gray-50">
      <div className="bg-white px-4 py-3 flex items-center border-b">
        <button onClick={onBack}><ChevronLeft className="w-5 h-5" /></button>
        <div className="flex-1 text-center font-medium">驾驶证认证</div>
        <div className="w-5" />
      </div>

      <div className="p-4 space-y-4">
        <div className="bg-indigo-50 border border-indigo-100 rounded-xl p-3 text-xs text-indigo-600">
          请上传有效的驾驶证照片，确保信息清晰可辨
        </div>

        <div>
          <div className="text-xs text-gray-500 mb-1.5">驾驶证编号</div>
          <input type="text" value={licenseNo} onChange={e => setLicenseNo(e.target.value)}
            placeholder="请输入驾驶证编号"
            className="w-full border border-gray-200 rounded-xl px-3 py-2.5 text-sm outline-none focus:border-emerald-500" />
        </div>

        <div>
          <div className="text-xs text-gray-500 mb-1.5">准驾车型</div>
          <div className="flex gap-2 flex-wrap">
            {["C1", "C2", "B1", "B2", "A1", "A2"].map(t => (
              <button key={t} onClick={() => setLicenseType(t)}
                className={`px-4 py-2 rounded-lg text-sm border ${licenseType === t ? "border-emerald-500 bg-emerald-50 text-emerald-600" : "border-gray-200 text-gray-600"}`}>
                {t}
              </button>
            ))}
          </div>
        </div>

        <PhotoUploader label="驾驶证照片" bizType="license" currentUrl={licenseUrl}
          onUploaded={setLicenseUrl} onToast={onToast} />

        <button onClick={handleSubmit} disabled={submitting}
          className="w-full bg-gradient-to-r from-emerald-500 to-teal-500 text-white py-3 rounded-full font-medium disabled:opacity-50">
          {submitting ? "提交中..." : "提交认证"}
        </button>
      </div>
    </div>
  );
}

/* ============================================================
   车辆认证页
   ============================================================ */
function DriverVehicleView({ driver, onBack, onToast }: any) {
  const [plateNo, setPlateNo] = useState(driver.plate || "");
  const [vehicleBrand, setVehicleBrand] = useState("");
  const [vehicleModel, setVehicleModel] = useState("");
  const [vehicleColor, setVehicleColor] = useState("");
  const [seatCount, setSeatCount] = useState(5);
  const [drivingLicenseUrl, setDrivingLicenseUrl] = useState("");
  const [vehiclePhotoUrl, setVehiclePhotoUrl] = useState("");
  const [submitting, setSubmitting] = useState(false);

  // 从后端已提交的数据回填
  useEffect(() => {
    if (driver.vehicleInfo) {
      if (driver.vehicleInfo.plate_no) setPlateNo(driver.vehicleInfo.plate_no);
      if (driver.vehicleInfo.vehicle_brand) setVehicleBrand(driver.vehicleInfo.vehicle_brand);
      if (driver.vehicleInfo.vehicle_model) setVehicleModel(driver.vehicleInfo.vehicle_model);
      if (driver.vehicleInfo.vehicle_color) setVehicleColor(driver.vehicleInfo.vehicle_color);
      if (driver.vehicleInfo.seat_count) setSeatCount(driver.vehicleInfo.seat_count);
      if (driver.vehicleInfo.driving_license_url) setDrivingLicenseUrl(driver.vehicleInfo.driving_license_url);
      if (driver.vehicleInfo.vehicle_photo_url) setVehiclePhotoUrl(driver.vehicleInfo.vehicle_photo_url);
    }
  }, []);

  const handleSubmit = async () => {
    if (!plateNo || !drivingLicenseUrl) {
      onToast("请填写车牌号并上传行驶证照片");
      return;
    }
    setSubmitting(true);
    const res = await updateVehicle({
      driver_id: 200000001,
      plate_no: plateNo,
      vehicle_brand: vehicleBrand,
      vehicle_model: vehicleModel,
      vehicle_color: vehicleColor,
      seat_count: seatCount,
      driving_license_url: drivingLicenseUrl,
      vehicle_photo_url: vehiclePhotoUrl,
    });
    setSubmitting(false);
    if (res?.success) {
      onToast("车辆信息提交成功，等待审核");
      onBack();
    } else {
      onToast(res?.message || "提交失败");
    }
  };

  return (
    <div className="min-h-full bg-gray-50">
      <div className="bg-white px-4 py-3 flex items-center border-b">
        <button onClick={onBack}><ChevronLeft className="w-5 h-5" /></button>
        <div className="flex-1 text-center font-medium">车辆认证</div>
        <div className="w-5" />
      </div>

      <div className="p-4 space-y-4">
        <div className="bg-violet-50 border border-violet-100 rounded-xl p-3 text-xs text-violet-600">
          请确保车辆信息与行驶证一致，审核通过后才能出车接单
        </div>

        <div>
          <div className="text-xs text-gray-500 mb-1.5">车牌号</div>
          <input type="text" value={plateNo} onChange={e => setPlateNo(e.target.value)}
            placeholder="例: 苏N·8F23K"
            className="w-full border border-gray-200 rounded-xl px-3 py-2.5 text-sm outline-none focus:border-emerald-500" />
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div>
            <div className="text-xs text-gray-500 mb-1.5">车辆品牌</div>
            <input type="text" value={vehicleBrand} onChange={e => setVehicleBrand(e.target.value)}
              placeholder="例: 日产"
              className="w-full border border-gray-200 rounded-xl px-3 py-2.5 text-sm outline-none focus:border-emerald-500" />
          </div>
          <div>
            <div className="text-xs text-gray-500 mb-1.5">车型</div>
            <input type="text" value={vehicleModel} onChange={e => setVehicleModel(e.target.value)}
              placeholder="例: 轩逸"
              className="w-full border border-gray-200 rounded-xl px-3 py-2.5 text-sm outline-none focus:border-emerald-500" />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div>
            <div className="text-xs text-gray-500 mb-1.5">车身颜色</div>
            <input type="text" value={vehicleColor} onChange={e => setVehicleColor(e.target.value)}
              placeholder="例: 银色"
              className="w-full border border-gray-200 rounded-xl px-3 py-2.5 text-sm outline-none focus:border-emerald-500" />
          </div>
          <div>
            <div className="text-xs text-gray-500 mb-1.5">座位数</div>
            <div className="flex gap-2">
              {[4, 5, 7].map(n => (
                <button key={n} onClick={() => setSeatCount(n)}
                  className={`flex-1 py-2 rounded-lg text-sm border ${seatCount === n ? "border-emerald-500 bg-emerald-50 text-emerald-600" : "border-gray-200 text-gray-600"}`}>
                  {n}座
                </button>
              ))}
            </div>
          </div>
        </div>

        <PhotoUploader label="行驶证照片" bizType="vehicle" currentUrl={drivingLicenseUrl}
          onUploaded={setDrivingLicenseUrl} onToast={onToast} />

        <PhotoUploader label="车辆外观照片" bizType="vehicle" currentUrl={vehiclePhotoUrl}
          onUploaded={setVehiclePhotoUrl} onToast={onToast} />

        <button onClick={handleSubmit} disabled={submitting}
          className="w-full bg-gradient-to-r from-emerald-500 to-teal-500 text-white py-3 rounded-full font-medium disabled:opacity-50">
          {submitting ? "提交中..." : "提交认证"}
        </button>
      </div>
    </div>
  );
}

/* ============================================================
   钱包概览页
   ============================================================ */
function DriverWalletView({ driver, onBack, onNav, onWithdraw, onToast }: any) {
  const [wallet, setWallet] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      setLoading(true);
      const data = await getWallet(200000001);
      setWallet(data);
      setLoading(false);
    }
    load();
  }, []);

  const balance = wallet?.balance ?? 0;
  const frozen = wallet?.frozen_amount ?? 0;

  return (
    <div className="min-h-full bg-gray-50">
      <div className="bg-white px-4 py-3 flex items-center border-b">
        <button onClick={onBack}><ChevronLeft className="w-5 h-5" /></button>
        <div className="flex-1 text-center font-medium">我的钱包</div>
        <button onClick={onWithdraw} className="text-emerald-600 text-sm">提现</button>
      </div>

      <div className="bg-gradient-to-r from-emerald-500 to-teal-500 p-5 text-white text-center">
        <div className="text-xs opacity-90 mb-1">可提现余额（元）</div>
        <div className="text-4xl font-light mb-3">¥{balance.toFixed(2)}</div>
        <div className="grid grid-cols-2 gap-3 text-[11px]">
          <div className="bg-white/15 rounded-xl py-2.5">
            <div className="font-medium">¥{frozen.toFixed(2)}</div>
            <div className="opacity-80 mt-0.5">冻结金额</div>
          </div>
          <div className="bg-white/15 rounded-xl py-2.5">
            <div className="font-medium">¥{(wallet?.total_income ?? 0).toFixed(2)}</div>
            <div className="opacity-80 mt-0.5">累计收入</div>
          </div>
        </div>
        {frozen > 0 && (
          <div className="text-[10px] opacity-80 mt-3 bg-white/10 rounded-lg px-3 py-1.5">
            冻结金额为T-3结算中的流水，3天后自动解冻可提现
          </div>
        )}
      </div>

      <div className="mx-3 mt-3 bg-white rounded-2xl p-4 shadow-sm">
        <div className="text-sm font-medium mb-3">收入概览</div>
        <div className="grid grid-cols-3 gap-3 text-center">
          {[
            { v: `¥${(wallet?.today_income ?? 0).toFixed(2)}`, l: "今日" },
            { v: `¥${(wallet?.week_income ?? 0).toFixed(2)}`, l: "本周" },
            { v: `¥${(wallet?.month_income ?? 0).toFixed(2)}`, l: "本月" },
          ].map(x => (
            <div key={x.l} className="bg-gray-50 rounded-xl py-3">
              <div className="text-base font-medium text-gray-900">{x.v}</div>
              <div className="text-[10px] text-gray-500 mt-0.5">{x.l}收入</div>
            </div>
          ))}
        </div>
      </div>

      <div className="mx-3 mt-3 bg-white rounded-2xl overflow-hidden shadow-sm">
        {[
          { icon: CreditCard, label: "银行卡管理", val: wallet?.has_bank_card ? wallet.bank_card_no : "未绑定", key: "bankcard", color: "text-blue-500" },
          { icon: Clock, label: "提现记录", val: "", key: "withdrawRecords", color: "text-violet-500" },
          { icon: TrendingUp, label: "收入明细", val: "", key: "income", color: "text-emerald-500" },
        ].map(({ icon: Icon, label, val, key, color }) => (
          <button key={label} onClick={() => key === "income" ? onNav("income") : onNav(key)}
            className="w-full flex items-center gap-3 px-4 py-3.5 border-b border-gray-50 last:border-0 hover:bg-gray-50">
            <div className={`w-9 h-9 rounded-full bg-gray-50 flex items-center justify-center ${color}`}>
              <Icon className="w-4 h-4" />
            </div>
            <span className="flex-1 text-left text-sm text-gray-900">{label}</span>
            {val && <span className="text-xs text-gray-400">{val}</span>}
            <ChevronRight className="w-4 h-4 text-gray-300" />
          </button>
        ))}
      </div>

      <div className="mx-3 mt-3 mb-6 bg-amber-50 border border-amber-200 rounded-2xl p-4">
        <div className="text-sm font-medium text-amber-700 mb-2">提现规则</div>
        <div className="space-y-1.5 text-xs text-amber-600">
          {[
            "每天最多提现3次，单笔上限¥5000",
            "T-3结算：冻结3天流水后可提现",
            "2小时内到账，节假日顺延",
            "更换银行卡每月限1次",
          ].map(t => (
            <div key={t} className="flex items-start gap-2">
              <CheckCircle2 className="w-3.5 h-3.5 mt-0.5 flex-shrink-0" />{t}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

/* ============================================================
   银行卡管理页
   ============================================================ */
function DriverBankCardView({ driver, onBack, onToast }: any) {
  const [cardInfo, setCardInfo] = useState<any>(null);
  const [editing, setEditing] = useState(false);
  const [bankName, setBankName] = useState("");
  const [bankCardNo, setBankCardNo] = useState("");
  const [accountName, setAccountName] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    async function load() {
      const data = await getBankCard(200000001);
      setCardInfo(data);
      if (data?.has_card) setEditing(false);
      else setEditing(true);
    }
    load();
  }, []);

  const handleSubmit = async () => {
    if (!bankName || !bankCardNo || !accountName) {
      onToast("银行名称、卡号、持卡人姓名不能为空");
      return;
    }
    if (bankCardNo.length > 20) {
      onToast("银行卡号不能超过20位");
      return;
    }
    setSubmitting(true);
    const params = {
      driver_id: 200000001,
      bank_name: bankName,
      bank_card_no: bankCardNo,
      account_name: accountName,
    };
    const res = cardInfo?.has_card
      ? await updateBankCard(params)
      : await bindBankCard(params);
    setSubmitting(false);
    if (res?.success) {
      onToast(cardInfo?.has_card ? "银行卡更换成功" : "银行卡绑定成功");
      const data = await getBankCard(200000001);
      setCardInfo(data);
      setEditing(false);
    } else {
      onToast(res?.message || "操作失败");
    }
  };

  const banks = ["工商银行", "建设银行", "农业银行", "中国银行", "交通银行", "招商银行", "邮储银行", "兴业银行", "浦发银行", "民生银行"];

  return (
    <div className="min-h-full bg-gray-50">
      <div className="bg-white px-4 py-3 flex items-center border-b">
        <button onClick={onBack}><ChevronLeft className="w-5 h-5" /></button>
        <div className="flex-1 text-center font-medium">银行卡管理</div>
        {cardInfo?.has_card && !editing && (
          <button onClick={() => setEditing(true)} className="text-emerald-600 text-sm">更换</button>
        )}
      </div>

      {cardInfo?.has_card && !editing ? (
        <div className="p-4">
          <div className="bg-gradient-to-br from-blue-600 to-indigo-700 rounded-2xl p-5 text-white relative overflow-hidden">
            <div className="text-lg font-bold mb-1">{cardInfo.bank_name}</div>
            <div className="text-2xl tracking-widest font-mono mb-4">{cardInfo.bank_card_no}</div>
            <div className="flex justify-between text-xs opacity-80">
              <span>{cardInfo.account_name}</span>
              <span>{cardInfo.card_type === 1 ? "借记卡" : "信用卡"}</span>
            </div>
            {cardInfo.branch_name && <div className="text-xs opacity-60 mt-1">{cardInfo.branch_name}</div>}
            <div className="absolute -right-4 -bottom-4 text-8xl opacity-10">💳</div>
          </div>
          <div className="text-xs text-amber-600 bg-amber-50 rounded-lg p-2.5 mt-3">
            更换银行卡每月限1次，上次修改时间：{cardInfo.last_modified_at || "无"}
          </div>
        </div>
      ) : (
        <div className="p-4 space-y-4">
          <div className="bg-blue-50 border border-blue-100 rounded-xl p-3 text-xs text-blue-600">
            {cardInfo?.has_card ? "更换银行卡每月限1次" : "绑定银行卡后即可提现"}，请确保信息准确
          </div>

          <div>
            <div className="text-xs text-gray-500 mb-1.5">银行名称</div>
            <div className="flex gap-2 flex-wrap">
              {banks.map(b => (
                <button key={b} onClick={() => setBankName(b)}
                  className={`px-3 py-1.5 rounded-lg text-xs border ${bankName === b ? "border-emerald-500 bg-emerald-50 text-emerald-600" : "border-gray-200 text-gray-600"}`}>
                  {b}
                </button>
              ))}
            </div>
          </div>

          <div>
            <div className="text-xs text-gray-500 mb-1.5">银行卡号</div>
            <input type="text" value={bankCardNo} onChange={e => setBankCardNo(e.target.value)}
              placeholder="请输入银行卡号" maxLength={20}
              className="w-full border border-gray-200 rounded-xl px-3 py-2.5 text-sm outline-none focus:border-emerald-500" />
          </div>

          <div>
            <div className="text-xs text-gray-500 mb-1.5">持卡人姓名</div>
            <input type="text" value={accountName} onChange={e => setAccountName(e.target.value)}
              placeholder="请输入持卡人姓名"
              className="w-full border border-gray-200 rounded-xl px-3 py-2.5 text-sm outline-none focus:border-emerald-500" />
          </div>

          <button onClick={handleSubmit} disabled={submitting}
            className="w-full bg-gradient-to-r from-emerald-500 to-teal-500 text-white py-3 rounded-full font-medium disabled:opacity-50">
            {submitting ? "提交中..." : cardInfo?.has_card ? "确认更换" : "确认绑定"}
          </button>
        </div>
      )}
    </div>
  );
}

/* ============================================================
   提现记录页
   ============================================================ */
function DriverWithdrawRecordsView({ onBack, onToast }: { onBack: () => void; onToast: (m: string) => void }) {
  const [records, setRecords] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      setLoading(true);
      const data = await getWithdrawRecords(200000001, 1, 50);
      if (data?.records) setRecords(data.records);
      setLoading(false);
    }
    load();
  }, []);

  const statusLabel: Record<number, { text: string; color: string }> = {
    1: { text: "处理中", color: "bg-amber-100 text-amber-700" },
    2: { text: "已到账", color: "bg-emerald-100 text-emerald-700" },
    3: { text: "失败", color: "bg-rose-100 text-rose-700" },
  };

  return (
    <div className="min-h-full bg-gray-50">
      <div className="bg-white px-4 py-3 flex items-center border-b">
        <button onClick={onBack}><ChevronLeft className="w-5 h-5" /></button>
        <div className="flex-1 text-center font-medium">提现记录</div>
        <div className="w-5" />
      </div>

      <div className="p-3 space-y-2">
        {records.map(r => {
          const st = statusLabel[r.status] || statusLabel[1];
          return (
            <div key={r.id} className="bg-white rounded-2xl p-4 shadow-sm">
              <div className="flex justify-between items-center mb-2">
                <span className="text-sm text-gray-800">{r.bank_name}</span>
                <span className="text-emerald-600 font-medium">-¥{r.amount.toFixed(2)}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-xs text-gray-400">{r.bank_card_no}</span>
                <span className={`text-[10px] px-2 py-0.5 rounded ${st.color}`}>{st.text}</span>
              </div>
              <div className="flex justify-between items-center mt-2 pt-2 border-t border-gray-50">
                <span className="text-[10px] text-gray-400">{r.apply_time}</span>
                {r.status === 2 && r.finish_time && (
                  <span className="text-[10px] text-gray-400">到账: {r.finish_time}</span>
                )}
                {r.status === 3 && r.fail_reason && (
                  <span className="text-[10px] text-rose-500">{r.fail_reason}</span>
                )}
              </div>
            </div>
          );
        })}
        {!loading && records.length === 0 && (
          <div className="text-center py-12 text-gray-400 text-sm">
            <div className="text-3xl mb-2">💰</div>暂无提现记录
          </div>
        )}
      </div>
    </div>
  );
}
