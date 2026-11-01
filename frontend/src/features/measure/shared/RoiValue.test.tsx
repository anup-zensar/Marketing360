import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { RoiValue } from './RoiValue';

describe('RoiValue (050 US3, FR-007, FR-008)', () => {
  it('names the model and the cost basis next to every ROI', () => {
    render(
      <>
        <RoiValue roi={50} model="U-shaped" basis="actual" />
        <RoiValue roi={null} model="U-shaped" basis="actual" />
        <RoiValue roi={-12} model="Last touch" basis="planned" />
      </>,
    );

    expect(screen.getByText('50%')).toBeInTheDocument();
    expect(screen.getByText('U-shaped · on actual cost')).toBeInTheDocument();
    expect(screen.getByText('No actual cost yet · U-shaped')).toBeInTheDocument();
    expect(screen.getByText('Last touch · on planned cost')).toBeInTheDocument();
  });
});
