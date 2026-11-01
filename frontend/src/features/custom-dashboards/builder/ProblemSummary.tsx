import { useEffect, useRef } from 'react';
import summaryStyles from '@/ui/FormErrorSummary/FormErrorSummary.module.css';
import type { Widget } from '../model';
import type { Problem } from './validate';
import { widgetName } from './widget-defaults';

export interface ProblemSummaryProps {
  problems: readonly Problem[];
  widgets: readonly Widget[];
  onGo: (problem: Problem) => void;
}

export function ProblemSummary({ problems, widgets, onGo }: ProblemSummaryProps) {
  const summary = useRef<HTMLDivElement>(null);

  useEffect(() => {
    summary.current?.focus();
  }, [problems]);

  if (problems.length === 0) {
    return null;
  }
  const title =
    problems.length === 1
      ? '1 problem needs your attention'
      : `${problems.length} problems need your attention`;

  function text(problem: Problem): string {
    const widget = widgets.find((each) => each.id === problem.widgetId);
    return widget ? `${widgetName(widget)}: ${problem.message}` : problem.message;
  }

  return (
    <div ref={summary} tabIndex={-1} className={summaryStyles.summary}>
      <div role="alert">
        <h2 className={summaryStyles.title}>{title}</h2>
        <ul className={summaryStyles.list}>
          {problems.map((problem) => (
            <li key={`${problem.field}-${problem.message}`}>
              <a
                href={`#${problem.field}`}
                onClick={(event) => {
                  event.preventDefault();
                  onGo(problem);
                }}
              >
                {text(problem)}
              </a>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}
