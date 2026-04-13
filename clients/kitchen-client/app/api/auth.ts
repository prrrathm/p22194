import { API_BASE, apiRequest } from "./client";

export interface TokenPair {
	access_token: string;
	refresh_token: string;
	expires_in: number;
}

export interface User {
	id: string;
	email: string;
	role: string;
}

async function publicPost<T>(path: string, body: unknown): Promise<T> {
	const res = await fetch(`${API_BASE}${path}`, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(body),
	});
	const data = await res.json();
	if (!res.ok) throw new Error(data.error ?? "Request failed");
	return data as T;
}

export function login(email: string, password: string): Promise<TokenPair> {
	return publicPost<TokenPair>("/auth/login", { email, password });
}

export function register(
	email: string,
	username: string,
	password: string,
): Promise<TokenPair> {
	return publicPost<TokenPair>("/auth/register", { email, username, password });
}

export function refreshToken(refresh_token: string): Promise<TokenPair> {
	return publicPost<TokenPair>("/auth/refresh", { refresh_token });
}

export function logout(refresh_token: string): Promise<void> {
	return apiRequest<void>("/auth/logout", {
		method: "POST",
		body: JSON.stringify({ refresh_token }),
	});
}

export function getMe(): Promise<User> {
	return apiRequest<User>("/auth/me");
}
