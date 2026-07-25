package auth

const (
	RouteRegister           = "/api/v1/auth/register"
	RouteLogin              = "/api/v1/auth/login"
	RouteLogout             = "/api/v1/auth/logout"
	RouteRefresh            = "/api/v1/auth/refresh"
	RouteVerifyEmail        = "/api/v1/auth/verify-email"
	RouteResendVerification = "/api/v1/auth/resend-verification"
	RouteForgotPassword     = "/api/v1/auth/forgot-password"
	RouteResetPassword      = "/api/v1/auth/reset-password"
	RouteChangePassword     = "/api/v1/auth/change-password"
)
