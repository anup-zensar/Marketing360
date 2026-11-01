import { useState } from 'react';
import { Button } from '@/ui/Button/Button';
import { Tabs } from '@/ui/Tabs/Tabs';
import { useDataSources } from '../charts';
import type { Widget } from '../model';
import type { BuilderAction, Draft } from './builder-reducer';
import { DashboardDetails } from './DashboardDetails';
import { DataTab } from './DataTab';
import { LayoutTab } from './LayoutTab';
import { StyleTab } from './StyleTab';
import type { Problem } from './validate';
import { problemAt } from './validate';
import { widgetName } from './widget-defaults';
import builder from './Builder.module.css';
import styles from './Inspector.module.css';

type Tab = 'data' | 'style' | 'layout';

export interface InspectorProps {
  draft: Draft;
  selected: Widget | null;
  problems: readonly Problem[];
  onAction: (action: BuilderAction) => void;
}

export function Inspector({ draft, selected, problems, onAction }: InspectorProps) {
  return (
    <aside aria-label="Inspector" className={builder.panel}>
      {selected ? (
        <WidgetInspector
          key={selected.id}
          draft={draft}
          widget={selected}
          problems={problems}
          onAction={onAction}
        />
      ) : (
        <DashboardDetails
          draft={draft}
          errorAt={(path) => problemAt(problems, path)}
          onAction={onAction}
        />
      )}
    </aside>
  );
}

interface WidgetInspectorProps {
  draft: Draft;
  widget: Widget;
  problems: readonly Problem[];
  onAction: (action: BuilderAction) => void;
}

function WidgetInspector({ draft, widget, problems, onAction }: WidgetInspectorProps) {
  const { id, query } = widget;
  const [tab, setTab] = useState<Tab>(query ? 'data' : 'style');
  const sources = useDataSources();
  const index = draft.widgets.findIndex((each) => each.id === id);
  const errorAt = (path: string) => problemAt(problems, `widgets[${index}].${path}`);
  const tabs = [
    ...(query ? [{ value: 'data' as const, label: 'Data' }] : []),
    { value: 'style' as const, label: 'Style' },
    { value: 'layout' as const, label: 'Layout' },
  ];

  return (
    <>
      <div className={styles.group}>
        <h2 className={styles.widgetTitle}>{widgetName(widget)}</h2>
        <Button size="sm" variant="tertiary" onClick={() => onAction({ type: 'select', id: null })}>
          Dashboard details
        </Button>
      </div>
      <Tabs label={`Settings of '${widgetName(widget)}'`} tabs={tabs} value={tab} onChange={setTab}>
        {tab === 'data' && query && (
          <DataTab
            widget={widget}
            query={query}
            sources={sources.data}
            sourcesFailed={sources.isError}
            errorAt={errorAt}
            onQuery={(next, typing) => onAction({ type: 'setQuery', id, query: next, typing })}
          />
        )}
        {tab === 'style' && (
          <StyleTab
            widget={widget}
            errorAt={errorAt}
            onDisplay={(display, typing) => onAction({ type: 'setDisplay', id, display, typing })}
            onText={(patch) => onAction({ type: 'setText', id, patch })}
          />
        )}
        {tab === 'layout' && (
          <LayoutTab
            key={JSON.stringify(widget.layout)}
            widget={widget}
            widgets={draft.widgets}
            onPlace={(layout) => onAction({ type: 'place', id, layout })}
          />
        )}
      </Tabs>
    </>
  );
}
