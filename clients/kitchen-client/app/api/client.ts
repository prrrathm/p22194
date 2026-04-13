import {
	clearTokens,
	getAccessToken,
	getRefreshToken,
	setTokens,
} from "~/utils/token";

export const API_BASE =
	typeof import.meta !== "undefined" && (import.meta as any).env?.VITE_API_URL
		? (import.meta as any).env.VITE_API_URL
		: "http://localhost:8080";

let isRefreshing = false;
let pendingRequests: Array<(token: string | null) => void> = [];

function notifyPending(token: string | null) {
	pendingRequests.forEach((cb) => cb(token));
	pendingRequests = [];
}

async function attemptRefresh(): Promise<string | null> {
	const refreshToken = getRefreshToken();
	if (!refreshToken) return null;

	const res = await fetch(`${API_BASE}/auth/refresh`, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ refresh_token: refreshToken }),
	});

	if (!res.ok) {
		clearTokens();
		return null;
	}

	const data = await res.json();
	setTokens(data.access_token, data.refresh_token);
	return data.access_token as string;
}

function buildHeaders(token: string | null, extra?: HeadersInit): HeadersInit {
	return {
		"Content-Type": "application/json",
		...(token ? { Authorization: `Bearer ${token}` } : {}),
		...extra,
	};
}

async function fetchOnce<T>(
	path: string,
	token: string | null,
	options: RequestInit,
): Promise<T> {
	const res = await fetch(`${API_BASE}${path}`, {
		...options,
		headers: buildHeaders(token, options.headers),
	});

	if (!res.ok) {
		const body = await res.json().catch(() => ({ error: "Request failed" }));
		throw new Error(body.error ?? "Request failed");
	}

	if (res.status === 204) return undefined as T;
	return res.json() as Promise<T>;
}

export async function apiRequest<T>(
	path: string,
	options: RequestInit = {},
): Promise<T> {
	const token = getAccessToken();

	const res = await fetch(`${API_BASE}${path}`, {
		...options,
		headers: buildHeaders(token, options.headers),
	});

	if (res.status !== 401) {
		if (!res.ok) {
			const body = await res.json().catch(() => ({ error: "Request failed" }));
			throw new Error(body.error ?? "Request failed");
		}
		if (res.status === 204) return undefined as T;
		return res.json() as Promise<T>;
	}

	// 401 — attempt token refresh
	if (!isRefreshing) {
		isRefreshing = true;
		const newToken = await attemptRefresh().catch(() => null);
		isRefreshing = false;
		notifyPending(newToken);

		if (!newToken) {
			if (typeof window !== "undefined") {
				window.location.href = "/auth/sign-in";
			}
			throw new Error("Session expired");
		}

		return fetchOnce<T>(path, newToken, options);
	}

	// Another refresh is already in flight — queue this request
	return new Promise<T>((resolve, reject) => {
		pendingRequests.push(async (newToken) => {
			if (!newToken) {
				reject(new Error("Session expired"));
				return;
			}
			try {
				resolve(await fetchOnce<T>(path, newToken, options));
			} catch (err) {
				reject(err);
			}
		});
	});
}
