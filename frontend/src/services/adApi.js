import {
  GetAdRuntime,
  OpenExternalURL as OpenAdExternalURL,
} from "@bindings/cursor/internal/bridge/adservice.js";

export function getAdRuntime() {
  return GetAdRuntime();
}

export function openAdExternalURL(url) {
  return OpenAdExternalURL(url);
}
