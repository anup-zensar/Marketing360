import { screen, waitFor } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { components } from '@/api/generated/schema';
import { testUser } from '@/testing/fixtures/users';
import { mockApi } from '@/testing/mock-api';
import { renderWithProviders } from '@/testing/render';
import { apiError, reply } from '@/testing/replies';
import { MediaCodesSection } from './MediaCodesSection';
import { taggedLink } from './tagged-link';

const AREA = 'GET /api/v1/activities/{activityId}/media-codes';
const CONFIG = 'GET /api/v1/performance/configuration';
const GENERATE = 'POST /api/v1/activities/{activityId}/media-codes';

type Area = components['schemas']['MediaCodeArea'];

function code(): components['schemas']['MediaCode'] {
  return {
    id: 'MC-1001',
    code: 9000001,
    activityId: 'A-1042',
    attributes: { Channel: 'paid-social' },
    landingPage: 'https://shell.com/spring?x=1',
    status: 'active',
    version: 1,
    createdAt: '2026-10-01T09:00:00Z',
    createdBy: { id: 'U-1001', name: 'Alex Morgan' },
    updatedAt: '2026-10-01T09:00:00Z',
    updatedBy: { id: 'U-1001', name: 'Alex Morgan' },
  };
}

function configuration(): components['schemas']['PerformanceConfiguration'] {
  return {
    activityTypeIds: ['AT-1'],
    codeAttributes: [
      { name: 'Channel', required: true },
      { name: 'Market', required: true },
    ],
    mediaCodeNext: 9000002,
    version: 1,
    createdAt: '2026-10-01T09:00:00Z',
    createdBy: { id: 'U-1001', name: 'Alex Morgan' },
    updatedAt: '2026-10-01T09:00:00Z',
    updatedBy: { id: 'U-1001', name: 'Alex Morgan' },
  };
}

function renderSection(area: Area, extra: Parameters<typeof mockApi>[0] = []) {
  mockApi([reply(AREA, { body: area }, { activityId: 'A-1042' }), ...extra]);
  return renderWithProviders(<MediaCodesSection activityId="A-1042" />, { user: testUser() });
}

describe('MediaCodesSection', () => {
  it('046 FR-040 stays hidden where the activity type does not use codes', async () => {
    renderSection({ requiresCodes: false, items: [] });

    expect(screen.getByText('Media codes')).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText('Media codes')).not.toBeInTheDocument());
  });

  it('046 FR-042 shows the code and its tagged link copy actions', async () => {
    renderSection({ requiresCodes: true, items: [code()] });

    expect(await screen.findByText('9000001')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Copy code' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Copy tagged link' })).toBeInTheDocument();
  });

  it('046 FR-041 reports every missing attribute at once in the generate dialog', async () => {
    const { user } = renderSection({ requiresCodes: true, items: [] }, [
      reply(CONFIG, { body: configuration() }),
      reply(
        GENERATE,
        {
          status: 400,
          body: apiError('validation_failed', 'Some fields need attention.', {
            fields: [
              { field: 'attributes.Channel', message: 'Give the code its Channel.' },
              { field: 'attributes.Market', message: 'Give the code its Market.' },
            ],
          }),
        },
        { activityId: 'A-1042' },
      ),
    ]);

    await user.click(await screen.findByRole('button', { name: 'Generate code' }));
    await user.click(await screen.findByRole('button', { name: 'Generate' }));

    expect(await screen.findByText('Give the code its Channel.')).toBeInTheDocument();
    expect(screen.getByText('Give the code its Market.')).toBeInTheDocument();
  });
});

describe('taggedLink', () => {
  it('046 FR-042 appends the code to the landing page with the right glue', () => {
    expect(taggedLink(code())).toBe('https://shell.com/spring?x=1&mc=9000001');
    expect(taggedLink({ ...code(), landingPage: 'https://shell.com/spring' })).toBe(
      'https://shell.com/spring?mc=9000001',
    );
    expect(taggedLink({ ...code(), landingPage: '' })).toBe('9000001');
  });
});
