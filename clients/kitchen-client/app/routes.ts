import {
	type RouteConfig,
	index,
	layout,
	route,
} from "@react-router/dev/routes";

export default [
	index("routes/home.tsx"),
	route("/auth/sign-in", "routes/auth/sign-in.tsx"),
	route("/auth/sign-up", "routes/auth/sign-up.tsx"),
	layout("routes/app/layout.tsx", [
		route("/app", "routes/app/index.tsx"),
		route("/app/doc/:id", "routes/app/document.$id.tsx"),
	]),
] satisfies RouteConfig;
