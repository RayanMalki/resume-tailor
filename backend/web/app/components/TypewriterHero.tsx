"use client";

import { useEffect, useMemo, useState } from "react";

type Phase = "typing" | "pausing" | "deleting";

export default function TypewriterHero({
  phrases,
  className
}: {
  phrases: string[];
  className?: string;
}) {
  const safePhrases = useMemo(
    () => phrases.filter((phrase) => phrase.trim().length > 0),
    [phrases]
  );
  const [index, setIndex] = useState(0);
  const [visible, setVisible] = useState("");
  const [phase, setPhase] = useState<Phase>("typing");

  useEffect(() => {
    if (safePhrases.length === 0) return;

    const current = safePhrases[index % safePhrases.length];
    let delay = 60;

    if (phase === "typing") {
      if (visible.length < current.length) {
        delay = 28 + Math.random() * 40;
        const timer = setTimeout(
          () => setVisible(current.slice(0, visible.length + 1)),
          delay
        );
        return () => clearTimeout(timer);
      }
      setPhase("pausing");
      return;
    }

    if (phase === "pausing") {
      const timer = setTimeout(() => setPhase("deleting"), 900);
      return () => clearTimeout(timer);
    }

    if (phase === "deleting") {
      if (visible.length > 0) {
        delay = 18 + Math.random() * 30;
        const timer = setTimeout(
          () => setVisible((prev) => prev.slice(0, -1)),
          delay
        );
        return () => clearTimeout(timer);
      }
      setPhase("typing");
      setIndex((prev) => (prev + 1) % safePhrases.length);
    }
  }, [safePhrases, index, visible, phase]);

  return (
    <div className={className}>
      <span>{visible}</span>
      <span className="ml-1 inline-block h-6 w-[2px] animate-caret bg-ember-400 align-middle" />
    </div>
  );
}
