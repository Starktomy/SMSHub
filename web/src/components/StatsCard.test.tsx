import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { StatCard } from '@/components/StatsCard';
import { Activity } from 'lucide-react';

describe('StatCard Component', () => {
  it('renders label and value correctly', () => {
    render(
      <StatCard
        label="测试指标"
        value={100}
        icon={Activity}
        colorClass="bg-blue-100 text-blue-600"
      />
    );
    expect(screen.getByText('测试指标')).toBeInTheDocument();
    expect(screen.getByText('100')).toBeInTheDocument();
  });

  it('renders unit when provided', () => {
    render(
      <StatCard
        label="测试指标"
        value={50}
        unit="%"
        icon={Activity}
        colorClass="bg-red-100"
      />
    );
    expect(screen.getByText('%')).toBeInTheDocument();
  });

  it('renders subValue correctly', () => {
    render(
      <StatCard
        label="测试指标"
        value={10}
        icon={Activity}
        colorClass="bg-blue-100"
        subValue="同比增加 5%"
      />
    );
    expect(screen.getByText('同比增加 5%')).toBeInTheDocument();
  });
});
