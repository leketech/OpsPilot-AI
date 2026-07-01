"use client";

interface PipelineStep {
  label: string;
  sub: string;
  duration: number;
}

interface Props {
  steps: PipelineStep[];
  activeStep: number; // index of the step currently running; steps.length = all done
}

export default function PipelineView({ steps, activeStep }: Props) {
  return (
    <div className="flex flex-col gap-2 fade-in">
      <div className="mb-4">
        <p className="text-sm font-semibold text-white">Agent Pipeline Running</p>
        <p className="text-xs text-slate-500 mt-0.5">
          {activeStep < steps.length
            ? `Step ${activeStep + 1} of ${steps.length}`
            : "Finalising…"}
        </p>
      </div>

      <div className="relative">
        {/* Vertical connector */}
        <div className="absolute left-[18px] top-8 bottom-8 w-px bg-slate-800" />

        <div className="space-y-3">
          {steps.map((step, i) => {
            const isDone   = i < activeStep;
            const isActive = i === activeStep;
            const isPending = i > activeStep;

            return (
              <div key={step.label} className="flex gap-4 items-start">
                {/* Status dot */}
                <div className="relative z-10 w-9 h-9 rounded-full border flex items-center justify-center shrink-0 transition-all duration-500"
                  style={{
                    borderColor: isDone   ? "#22c55e"
                               : isActive ? "#22d3ee"
                               : "#334155",
                    background:  isDone   ? "rgba(34,197,94,0.1)"
                               : isActive ? "rgba(34,211,238,0.1)"
                               : "transparent",
                  }}
                >
                  {isDone ? (
                    <svg className="w-4 h-4 text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                    </svg>
                  ) : isActive ? (
                    <span className="w-2.5 h-2.5 rounded-full bg-cyan-400 step-pulse" />
                  ) : (
                    <span className="w-2 h-2 rounded-full bg-slate-700" />
                  )}
                </div>

                {/* Step content */}
                <div className={`flex-1 border rounded-lg p-3.5 transition-all duration-500 ${
                  isDone    ? "border-green-900/40 bg-green-950/10"
                  : isActive ? "border-cyan-800/60 bg-cyan-950/20"
                  : "border-slate-800 bg-transparent opacity-40"
                }`}>
                  <p className={`text-sm font-semibold transition-colors ${
                    isDone    ? "text-green-400"
                    : isActive ? "text-cyan-300"
                    : "text-slate-500"
                  }`}>
                    {step.label}
                  </p>
                  <p className="text-xs text-slate-500 mt-0.5 font-mono">{step.sub}</p>

                  {/* Scanning bar for active step — keyframe defined in globals.css */}
                  {isActive && (
                    <div className="mt-2.5 h-0.5 bg-slate-800 rounded-full overflow-hidden">
                      <div
                        className="h-full bg-cyan-500 rounded-full"
                        style={{ animationName: "scan", animationDuration: `${step.duration}ms`, animationTimingFunction: "linear", animationFillMode: "forwards" }}
                      />
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
