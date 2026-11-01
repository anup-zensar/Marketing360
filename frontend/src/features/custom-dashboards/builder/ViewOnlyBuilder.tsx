import { useNavigate } from 'react-router';
import { asApiError } from '@/api/errors';
import { Button, LinkButton } from '@/ui/Button/Button';
import { FailureMessage } from '@/ui/FormPanel/FormPanel';
import { PageHeader } from '@/ui/PageHeader/PageHeader';
import { useToast } from '@/ui/Toast/useToast';
import { usePageTitle } from '@/ui/usePageTitle';
import { WidgetView } from '../charts';
import type { CustomDashboard } from '../model';
import { useCopyDashboard } from './api';
import { gridArea } from './canvas-dom';
import { widgetName } from './widget-defaults';
import styles from './Builder.module.css';

export interface ViewOnlyBuilderProps {
  dashboard: CustomDashboard;
}

export function ViewOnlyBuilder({ dashboard }: ViewOnlyBuilderProps) {
  const navigate = useNavigate();
  const { show } = useToast();
  const copy = useCopyDashboard();
  const canCopy = dashboard.access.copy.available;
  usePageTitle(dashboard.name);

  function makeCopy() {
    if (copy.isPending) {
      return;
    }
    copy.mutate(dashboard.id, {
      onSuccess: (copied) => {
        show({ message: 'Copy made' });
        void navigate(`/builder/${copied.id}`);
      },
    });
  }

  return (
    <>
      <PageHeader
        breadcrumb={[{ label: 'Dashboards', to: '/dashboards' }, { label: dashboard.name }]}
        title={dashboard.name}
        description={`Only ${dashboard.owner.name} can edit this dashboard.`}
        actions={
          <>
            <LinkButton to={`/dashboards/${dashboard.id}`}>View</LinkButton>
            {canCopy && (
              <Button variant="primary" aria-busy={copy.isPending} onClick={makeCopy}>
                Make a copy
              </Button>
            )}
          </>
        }
      />
      {copy.error && <FailureMessage failure={{ message: asApiError(copy.error).message }} />}
      <div className={styles.canvas}>
        <div className={styles.grid}>
          {dashboard.widgets.map((widget) => (
            <section
              key={widget.id}
              aria-label={widgetName(widget)}
              className={styles.widget}
              style={gridArea(widget.layout)}
            >
              <div className={styles.preview}>
                <WidgetView
                  widget={widget}
                  filters={dashboard.filters}
                  mode="view"
                  dashboardName={dashboard.name}
                  viewerOwnsDashboard={false}
                />
              </div>
            </section>
          ))}
        </div>
      </div>
    </>
  );
}
