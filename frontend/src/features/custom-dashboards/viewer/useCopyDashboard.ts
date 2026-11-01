import { useNavigate } from 'react-router';
import { ApiError, isSessionLost } from '@/api/errors';
import { useToast } from '@/ui/Toast/useToast';
import { useCopyCustomDashboard } from './api';

export function useCopyDashboard() {
  const copy = useCopyCustomDashboard();
  const navigate = useNavigate();
  const { show } = useToast();

  async function copyAndOpen(dashboardId: string) {
    if (copy.isPending) {
      return;
    }
    try {
      const made = await copy.mutateAsync(dashboardId);
      show({ message: 'Copy made' });
      void navigate(`/builder/${encodeURIComponent(made.id)}`);
    } catch (error) {
      const failure = error instanceof ApiError ? error : new ApiError('other');
      if (!isSessionLost(failure)) {
        show({ message: failure.message });
      }
    }
  }

  return { copyAndOpen, isCopying: copy.isPending };
}
