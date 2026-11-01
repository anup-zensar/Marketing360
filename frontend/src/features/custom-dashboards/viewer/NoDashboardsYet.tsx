import { ChartColumn, ChartPie, FileSpreadsheet, Plus } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { useId } from 'react';
import { LinkButton } from '@/ui/Button/Button';
import styles from './NoDashboardsYet.module.css';

interface Idea {
  icon: LucideIcon;
  title: string;
  text: string;
}

const IDEAS: Idea[] = [
  {
    icon: ChartColumn,
    title: 'Planned cost by month',
    text: 'Columns of your plans’ planned cost, month by month.',
  },
  {
    icon: ChartPie,
    title: 'Budget by campaign',
    text: 'A donut of where each campaign’s budget goes.',
  },
  {
    icon: FileSpreadsheet,
    title: 'Leads from your own file',
    text: 'Upload a CSV or Excel file and chart it next to your plans.',
  },
];

export function NoDashboardsYet({ canBuild }: { canBuild: boolean }) {
  const ideasId = useId();

  if (!canBuild) {
    return (
      <div className={styles.empty}>
        <DashboardSketch />
        <p className={styles.title}>You have no dashboards of your own</p>
        <p className={styles.text}>Dashboards your teams share with you appear below.</p>
      </div>
    );
  }

  return (
    <div className={styles.empty}>
      <DashboardSketch />
      <p className={styles.title}>You haven&apos;t created any dashboards yet</p>
      <p className={styles.text}>
        Would you like to create one? Pick a chart, choose your data and arrange it the way you like
        — no code needed.
      </p>
      <h3 id={ideasId} className={styles.ideasTitle}>
        Ideas to start with
      </h3>
      <ul role="list" aria-labelledby={ideasId} className={styles.ideas}>
        {IDEAS.map(({ icon: Icon, title, text }) => (
          <li key={title} className={styles.idea}>
            <Icon aria-hidden className={styles.ideaIcon} />
            <span className={styles.ideaTitle}>{title}</span>
            <span className={styles.ideaText}>{text}</span>
          </li>
        ))}
      </ul>
      <div className={styles.actions}>
        <LinkButton to="/builder" icon={<Plus />}>
          Create your first dashboard
        </LinkButton>
        <LinkButton to="/datasets" variant="tertiary">
          Upload your own data
        </LinkButton>
      </div>
    </div>
  );
}

function DashboardSketch() {
  return (
    <svg className={styles.sketch} viewBox="0 0 240 132" aria-hidden focusable="false">
      <rect className={styles.panel} x="1" y="1" width="238" height="130" rx="10" />
      <rect className={styles.panel} x="14" y="14" width="102" height="104" rx="6" />
      {[34, 58, 46, 76, 64].map((height, index) => (
        <rect
          key={height}
          className={styles[`bar${(index % 3) + 1}`]}
          x={26 + index * 17}
          y={106 - height}
          width="11"
          height={height}
          rx="2"
        />
      ))}
      <rect className={styles.panel} x="126" y="14" width="100" height="46" rx="6" />
      <polyline className={styles.line} points="136,48 152,38 168,42 184,28 200,32 216,22" />
      <rect className={styles.panel} x="126" y="70" width="100" height="48" rx="6" />
      <circle className={styles.donutBase} cx="152" cy="94" r="14" />
      <circle className={styles.donutPart} cx="152" cy="94" r="14" pathLength="100" />
      <rect className={styles.kpi} x="176" y="86" width="38" height="7" rx="3" />
      <rect className={styles.kpiFaint} x="176" y="98" width="26" height="5" rx="2" />
    </svg>
  );
}
