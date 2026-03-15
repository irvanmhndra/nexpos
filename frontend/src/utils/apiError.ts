interface ApiErrorResponse {
  message?: string;
}

interface AxiosLikeError {
  response?: {
    data?: ApiErrorResponse;
  };
}

export function getApiErrorMessage(error: unknown, fallback = 'Terjadi kesalahan, coba lagi'): string {
  const axiosError = error as AxiosLikeError;
  return axiosError?.response?.data?.message ?? fallback;
}
