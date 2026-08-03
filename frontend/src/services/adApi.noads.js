export function getAdRuntime() {
  return Promise.reject(new Error("广告功能已禁用"));
}

export function openAdExternalURL() {
  return Promise.reject(new Error("广告功能已禁用"));
}
