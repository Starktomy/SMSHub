import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { SignalStrength } from '@/components/SignalStrength';

describe('SignalStrength Component', () => {
  it('renders correct signal level description', () => {
    render(<SignalStrength level={31} showText={true} />);
    expect(screen.getByText(/优秀/)).toBeInTheDocument();
    expect(screen.getByText(/100%/)).toBeInTheDocument();
  });

  it('renders poor signal correctly', () => {
    render(<SignalStrength level={5} showText={true} />);
    expect(screen.getByText(/较差/)).toBeInTheDocument();
  });

  it('renders zero signal correctly', () => {
    render(<SignalStrength level={0} showText={true} />);
    expect(screen.getByText(/无信号/)).toBeInTheDocument();
  });

  it('applies custom className', () => {
    const { container } = render(<SignalStrength level={20} className="custom-class" />);
    expect(container.firstChild).toHaveClass('custom-class');
  });
});
