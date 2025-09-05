// Lightweight animation utilities to replace Framer Motion
// This provides similar functionality with much smaller bundle size

export interface AnimationProps {
  children: React.ReactNode;
  className?: string;
  style?: React.CSSProperties;
  onClick?: () => void;
  onMouseEnter?: () => void;
  onMouseLeave?: () => void;
}

// Animation variants
export const fadeIn = {
  initial: { opacity: 0 },
  animate: { opacity: 1 },
  transition: { duration: 0.5 }
};

export const slideUp = {
  initial: { opacity: 0, transform: 'translateY(20px)' },
  animate: { opacity: 1, transform: 'translateY(0)' },
  transition: { duration: 0.6 }
};

export const slideInLeft = {
  initial: { opacity: 0, transform: 'translateX(-30px)' },
  animate: { opacity: 1, transform: 'translateX(0)' },
  transition: { duration: 0.6 }
};

export const slideInRight = {
  initial: { opacity: 0, transform: 'translateX(30px)' },
  animate: { opacity: 1, transform: 'translateX(0)' },
  transition: { duration: 0.6 }
};

export const scaleIn = {
  initial: { opacity: 0, transform: 'scale(0.8)' },
  animate: { opacity: 1, transform: 'scale(1)' },
  transition: { duration: 0.5 }
};

export const staggerContainer = {
  animate: {
    transition: {
      staggerChildren: 0.1
    }
  }
};

export const staggerItem = {
  initial: { opacity: 0, y: 20 },
  animate: { opacity: 1, y: 0 }
};

// Lightweight motion components
export const MotionDiv: React.FC<AnimationProps & { 
  initial?: any; 
  animate?: any; 
  transition?: any;
  whileHover?: any;
  whileTap?: any;
}> = ({ 
  children, 
  className = '', 
  style = {}, 
  initial, 
  animate, 
  transition,
  whileHover,
  whileTap,
  onClick,
  onMouseEnter,
  onMouseLeave,
  ...props 
}) => {
  const [isVisible, setIsVisible] = React.useState(false);
  const [isHovered, setIsHovered] = React.useState(false);
  const [isTapped, setIsTapped] = React.useState(false);
  const ref = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsVisible(true);
        }
      },
      { threshold: 0.1 }
    );

    if (ref.current) {
      observer.observe(ref.current);
    }

    return () => observer.disconnect();
  }, []);

  const getAnimationStyle = () => {
    let animationStyle = { ...style };
    
    if (initial && !isVisible) {
      animationStyle = { ...animationStyle, ...initial };
    } else if (animate && isVisible) {
      animationStyle = { ...animationStyle, ...animate };
    }

    if (whileHover && isHovered) {
      animationStyle = { ...animationStyle, ...whileHover };
    }

    if (whileTap && isTapped) {
      animationStyle = { ...animationStyle, ...whileTap };
    }

    return animationStyle;
  };

  return (
    <div
      ref={ref}
      className={`transition-all duration-500 ease-out ${className}`}
      style={getAnimationStyle()}
      onClick={onClick}
      onMouseEnter={(e) => {
        setIsHovered(true);
        onMouseEnter?.(e);
      }}
      onMouseLeave={(e) => {
        setIsHovered(false);
        onMouseLeave?.(e);
      }}
      onMouseDown={() => setIsTapped(true)}
      onMouseUp={() => setIsTapped(false)}
      {...props}
    >
      {children}
    </div>
  );
};

export const MotionButton: React.FC<AnimationProps & { 
  initial?: any; 
  animate?: any; 
  transition?: any;
  whileHover?: any;
  whileTap?: any;
}> = ({ 
  children, 
  className = '', 
  style = {}, 
  initial, 
  animate, 
  transition,
  whileHover,
  whileTap,
  onClick,
  ...props 
}) => {
  const [isVisible, setIsVisible] = React.useState(false);
  const [isHovered, setIsHovered] = React.useState(false);
  const [isTapped, setIsTapped] = React.useState(false);
  const ref = React.useRef<HTMLButtonElement>(null);

  React.useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsVisible(true);
        }
      },
      { threshold: 0.1 }
    );

    if (ref.current) {
      observer.observe(ref.current);
    }

    return () => observer.disconnect();
  }, []);

  const getAnimationStyle = () => {
    let animationStyle = { ...style };
    
    if (initial && !isVisible) {
      animationStyle = { ...animationStyle, ...initial };
    } else if (animate && isVisible) {
      animationStyle = { ...animationStyle, ...animate };
    }

    if (whileHover && isHovered) {
      animationStyle = { ...animationStyle, ...whileHover };
    }

    if (whileTap && isTapped) {
      animationStyle = { ...animationStyle, ...whileTap };
    }

    return animationStyle;
  };

  return (
    <button
      ref={ref}
      className={`transition-all duration-300 ease-out ${className}`}
      style={getAnimationStyle()}
      onClick={onClick}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      onMouseDown={() => setIsTapped(true)}
      onMouseUp={() => setIsTapped(false)}
      {...props}
    >
      {children}
    </button>
  );
};

export const MotionSection: React.FC<AnimationProps & { 
  initial?: any; 
  animate?: any; 
  transition?: any;
}> = ({ 
  children, 
  className = '', 
  style = {}, 
  initial, 
  animate, 
  transition,
  ...props 
}) => {
  const [isVisible, setIsVisible] = React.useState(false);
  const ref = React.useRef<HTMLElement>(null);

  React.useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsVisible(true);
        }
      },
      { threshold: 0.1 }
    );

    if (ref.current) {
      observer.observe(ref.current);
    }

    return () => observer.disconnect();
  }, []);

  const getAnimationStyle = () => {
    let animationStyle = { ...style };
    
    if (initial && !isVisible) {
      animationStyle = { ...animationStyle, ...initial };
    } else if (animate && isVisible) {
      animationStyle = { ...animationStyle, ...animate };
    }

    return animationStyle;
  };

  return (
    <section
      ref={ref}
      className={`transition-all duration-700 ease-out ${className}`}
      style={getAnimationStyle()}
      {...props}
    >
      {children}
    </section>
  );
};

// Hook for intersection observer
export const useInView = (options = { threshold: 0.1 }) => {
  const [isInView, setIsInView] = React.useState(false);
  const ref = React.useRef<HTMLElement>(null);

  React.useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        setIsInView(entry.isIntersecting);
      },
      options
    );

    if (ref.current) {
      observer.observe(ref.current);
    }

    return () => observer.disconnect();
  }, [options]);

  return { ref, isInView };
};

// AnimatePresence replacement
export const AnimatePresence: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return <>{children}</>;
};

// Import React for the components
import React from 'react';
