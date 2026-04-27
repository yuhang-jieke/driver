import AMapLoader from "@amap/amap-jsapi-loader";

const AMap_KEY = "06889c89297fbaa64fd225235bacc46f";
const AMap_SECURITY_CODE = "1a793a03f4e64cab00bed25e1aab3069";

(window as any)._AMapSecurity = { securityJsCode: AMap_SECURITY_CODE };

let _AMap: any = null;
let _loadPromise: Promise<any> | null = null;

export async function loadAMap() {
  if (_AMap) return _AMap;
  if (!_loadPromise) {
    _loadPromise = AMapLoader.load({
      key: AMap_KEY,
      version: "2.0",
      plugins: ["AMap.Geocoder", "AMap.Driving", "AMap.AutoComplete", "AMap.Geolocation"],
    });
    _loadPromise.then((AMap) => { _AMap = AMap; });
  }
  return _loadPromise;
}

export function createMap(container: HTMLElement, opts?: any) {
  const AMap = _AMap;
  return new AMap.Map(container, {
    zoom: 13,
    center: [118.3, 33.95],
    viewMode: "2D",
    ...opts,
  });
}

// 坐标格式统一为 [lat, lng] 传入，内部转换为 [lng, lat]
export async function geocode(address: string): Promise<[number, number] | null> {
  const AMap = await loadAMap();
  return new Promise((resolve) => {
    const geocoder = new AMap.Geocoder({ city: "宿迁市" });
    geocoder.getLocation(address, (status: string, result: any) => {
      if (status === "complete" && result.geocodes?.length > 0) {
        const loc = result.geocodes[0].location;
        resolve([loc.getLat(), loc.getLng()]);
      } else {
        resolve(null);
      }
    });
  });
}

export async function reverseGeocode(lat: number, lng: number): Promise<string> {
  const AMap = await loadAMap();
  return new Promise((resolve) => {
    const geocoder = new AMap.Geocoder({ radius: 1000, extensions: "base" });
    geocoder.getAddress([lng, lat], (status: string, result: any) => {
      if (status === "complete" && result.regeocode) {
        resolve(result.regeocode.formattedAddress || result.regeocode.sematicDescription || "");
      } else {
        resolve("");
      }
    });
  });
}

export async function searchDrivingRoute(origin: [number, number], destination: [number, number]) {
  const AMap = await loadAMap();
  return new Promise<{ distance: number; time: number; paths: any[] }>((resolve) => {
    const driving = new AMap.Driving({ map: null, panel: null });
    driving.search(
      new AMap.LngLat(origin[1], origin[0]), // input [lat, lng] -> AMap expects [lng, lat]
      new AMap.LngLat(destination[1], destination[0]),
      {},
      (status: string, result: any) => {
        if (status === "complete" && result.routes?.length > 0) {
          resolve(result.routes[0]);
        } else {
          resolve({ distance: 0, time: 0, paths: [] });
        }
      }
    );
  });
}