import { Signal, SignalHigh, SignalLow, SignalMedium, SignalZero } from "lucide-react";
import { cn } from "@/lib/utils";

interface SignalStrengthProps {
  level: number; // 0-31
  className?: string;
  showText?: boolean;
}

export function SignalStrength({ level, className, showText = false }: SignalStrengthProps) {
  // Convert CSQ (0-31) to percentage
  const percentage = Math.round((level / 31) * 100);

  // Determine icon and color based on signal level
  let Icon = SignalZero;
  let colorClass = "text-gray-400";
  let text = "无信号";

  if (level >= 20) {
    Icon = Signal;
    colorClass = "text-green-500";
    text = "优秀";
  } else if (level >= 15) {
    Icon = SignalHigh;
    colorClass = "text-blue-500";
    text = "良好";
  } else if (level >= 10) {
    Icon = SignalMedium;
    colorClass = "text-yellow-500";
    text = "一般";
  } else if (level > 0) {
    Icon = SignalLow;
    colorClass = "text-red-500";
    text = "较差";
  }

  return (
    <div className={cn("flex items-center gap-1", className)} title={`信号强度: ${percentage}% (${level})`}>
      <Icon className={cn("w-4 h-4", colorClass)} />
      {showText && <span className="text-xs text-gray-500">{text} ({percentage}%)</span>}
    </div>
  );
}
