import { Link } from "react-router";
import { SignInForm } from "~/components/auth/sign-in-form";

export default function SignInPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-neutral-50 dark:bg-neutral-950 px-4">
      <div className="w-full max-w-sm">
        <div className="mb-8 text-center">
          <h1 className="text-2xl font-bold text-neutral-900 dark:text-neutral-100">
            Welcome back
          </h1>
          <p className="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
            Sign in to your workspace
          </p>
        </div>

        <div className="rounded-2xl border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-900 p-8 shadow-sm">
          <SignInForm />
        </div>

        <p className="mt-6 text-center text-sm text-neutral-500 dark:text-neutral-400">
          Don't have an account?{" "}
          <Link
            to="/auth/sign-up"
            className="font-medium text-neutral-900 dark:text-neutral-100 underline underline-offset-4 hover:no-underline"
          >
            Sign up
          </Link>
        </p>
      </div>
    </div>
  );
}
