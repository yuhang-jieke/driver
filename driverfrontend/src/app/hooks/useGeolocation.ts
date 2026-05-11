import { useState, useEffect, useCallback, useRef } from "react";

interface GeoState {
  lat: number;
  lng: number;
  address: string;
  accuracy: number | null;
  error: string | null;
  loading: boolean;
  source: 'gps' | 'ip' | 'static';
}

const DEFAULT_CENTER: [number, number] = [33.95, 118.3];

const IP_SERVICES = [
  'https://freegeoip.app/json/',
  'https://ipwho.is/',
  'https://ipinfo.io/json',
];

async function fallbackIpLocation(): Promise<{ lat: number; lng: number; ip: string } | null> {
  for (const url of IP_SERVICES) {
    try {
      const res = await fetch(url, { signal: AbortSignal.timeout(3000) });
      const data = await res.json();
      
      if (url.includes('freegeoip')) {
        if (data.latitude && data.longitude) {
          console.log(`[定位] IP定位降级: ${data.latitude}, ${data.longitude}`);
          return { lat: data.latitude, lng: data.longitude, ip: data.ip || '' };
        }
      } else if (url.includes('ipwho')) {
        if (data.latitude && data.longitude) {
          console.log(`[定位] IP定位降级: ${data.latitude}, ${data.longitude}`);
          if (data.ip) console.log(`[定位] IP定位使用的地址: ${data.ip}`);
          return { lat: data.latitude, lng: data.longitude, ip: data.ip || '' };
        }
      } else if (url.includes('ipinfo')) {
        if (data.loc) {
          const [lat, lng] = data.loc.split(',');
          console.log(`[定位] IP定位降级: ${lat}, ${lng}`);
          if (data.ip) console.log(`[定位] IP定位使用的地址: ${data.ip}`);
          return { lat: parseFloat(lat), lng: parseFloat(lng), ip: data.ip || '' };
        }
      }
    } catch { continue; }
  }
  return null;
}

export function useGeolocation() {
  const [state, setState] = useState<GeoState>({
    lat: DEFAULT_CENTER[0], lng: DEFAULT_CENTER[1], address: "", accuracy: null, error: null, loading: true, source: 'static',
  });
  const mountedRef = useRef(true);

  const start = useCallback(() => {
    if (!navigator.geolocation) {
      console.warn("[定位] 不支持 Geolocation API，切换至IP定位");
      fallbackIpLocation().then((result) => {
        if (mountedRef.current && result) {
          setState({ lat: result.lat, lng: result.lng, address: "IP定位 (网络精度)", accuracy: 5000, error: null, loading: false, source: 'ip' });
        }
      });
      return;
    }

    const onSuccess = (pos: GeolocationPosition) => {
      if (!mountedRef.current) return;
      console.log(`[定位] 成功: ${pos.coords.latitude.toFixed(4)}, ${pos.coords.longitude.toFixed(4)} (精度: ±${Math.round(pos.coords.accuracy)}m)`);
      setState((s) => ({ ...s, lat: pos.coords.latitude, lng: pos.coords.longitude, accuracy: pos.coords.accuracy, loading: false, error: null, source: 'gps' }));
    };

    const onError = (err: GeolocationPositionError) => {
      if (!mountedRef.current) return;
      console.warn(`[定位] 失败 [${err.code}]: ${err.message}，切换至IP定位...`);
      
      fallbackIpLocation().then((result) => {
        if (mountedRef.current) {
          if (result) {
            setState((s) => ({ ...s, lat: result.lat, lng: result.lng, address: "IP定位 (网络精度)", accuracy: 5000, error: null, loading: false, source: 'ip' }));
          } else {
            const msgs: Record<number, string> = { 1: "请允许浏览器获取位置权限", 2: "定位信息不可用", 3: "定位请求超时" };
            setState((s) => ({ ...s, loading: false, error: msgs[err.code] || "无法获取位置", source: 'static' }));
          }
        }
      });
    };

    setState((s) => ({ ...s, loading: true, error: null }));
    navigator.geolocation.getCurrentPosition(onSuccess, onError, {
      enableHighAccuracy: true,
      timeout: 8000,
      maximumAge: 60000,
    });
  }, []);

  useEffect(() => {
    mountedRef.current = true;
    start();
    return () => { mountedRef.current = false; };
  }, [start]);

  return { ...state, refresh: start };
}