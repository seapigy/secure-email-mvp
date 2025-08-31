/**
 * Logging utility for the Secure Email MVP
 * Replaces console.log/error statements with a centralized logging system
 */

export enum LogLevel {
  DEBUG = 0,
  INFO = 1,
  WARN = 2,
  ERROR = 3,
  CRITICAL = 4
}

interface LogEntry {
  timestamp: string;
  level: LogLevel;
  message: string;
  data?: unknown;
  component?: string;
  userId?: string;
}

class Logger {
  private logLevel: LogLevel;
  private logs: LogEntry[] = [];
  private maxLogs = 1000; // Prevent memory leaks

  constructor(level: LogLevel = LogLevel.INFO) {
    this.logLevel = level;
  }

  private shouldLog(level: LogLevel): boolean {
    return level >= this.logLevel;
  }

  private addLog(level: LogLevel, message: string, data?: unknown, component?: string): void {
    if (!this.shouldLog(level)) return;

    const entry: LogEntry = {
      timestamp: new Date().toISOString(),
      level,
      message,
      data,
      component
    };

    this.logs.push(entry);

    // Prevent memory leaks by limiting log entries
    if (this.logs.length > this.maxLogs) {
      this.logs = this.logs.slice(-this.maxLogs / 2);
    }

    // In development, still use console for debugging
    if (typeof window !== 'undefined' && window.location.hostname === 'localhost') {
      const consoleMethod = this.getConsoleMethod(level);
      consoleMethod(`[${entry.component || 'APP'}] ${message}`, data || '');
    }

    // TODO: In production, send to logging service
    // this.sendToLoggingService(entry);
  }

  private getConsoleMethod(level: LogLevel): (...args: unknown[]) => void {
    switch (level) {
      case LogLevel.DEBUG:
        return console.debug;
      case LogLevel.INFO:
        return console.info;
      case LogLevel.WARN:
        return console.warn;
      case LogLevel.ERROR:
      case LogLevel.CRITICAL:
        return console.error;
      default:
        return console.log;
    }
  }

  debug(message: string, data?: unknown, component?: string): void {
    this.addLog(LogLevel.DEBUG, message, data, component);
  }

  info(message: string, data?: unknown, component?: string): void {
    this.addLog(LogLevel.INFO, message, data, component);
  }

  warn(message: string, data?: unknown, component?: string): void {
    this.addLog(LogLevel.WARN, message, data, component);
  }

  error(message: string, data?: unknown, component?: string): void {
    this.addLog(LogLevel.ERROR, message, data, component);
  }

  critical(message: string, data?: unknown, component?: string): void {
    this.addLog(LogLevel.CRITICAL, message, data, component);
  }

  // Performance logging
  performance(component: string, operation: string, duration: number, metadata?: unknown): void {
    this.info(`Performance: ${component}.${operation} took ${duration.toFixed(2)}ms`, metadata, 'PERFORMANCE');
  }

  // Security logging
  security(event: string, details: unknown, component?: string): void {
    this.info(`[SECURITY] ${event}`, details, component || 'SECURITY');
  }

  // API logging
  api(method: string, url: string, status: number, duration: number, error?: unknown): void {
    const level = error ? LogLevel.ERROR : LogLevel.INFO;
    const message = `${method} ${url} - ${status} (${duration.toFixed(2)}ms)`;
    this.addLog(level, message, { error }, 'API');
  }

  // Get logs for debugging
  getLogs(): LogEntry[] {
    return [...this.logs];
  }

  // Clear logs
  clear(): void {
    this.logs = [];
  }

  // Export logs
  export(): string {
    return JSON.stringify(this.logs, null, 2);
  }
}

// Create singleton instance
export const logger = new Logger(
  typeof window !== 'undefined' && window.location.hostname === 'localhost' ? LogLevel.DEBUG : LogLevel.INFO
);

// Convenience functions
export const log = {
  debug: (message: string, data?: unknown, component?: string) => logger.debug(message, data, component),
  info: (message: string, data?: unknown, component?: string) => logger.info(message, data, component),
  warn: (message: string, data?: unknown, component?: string) => logger.warn(message, data, component),
  error: (message: string, data?: unknown, component?: string) => logger.error(message, data, component),
  critical: (message: string, data?: unknown, component?: string) => logger.critical(message, data, component),
  performance: (component: string, operation: string, duration: number, metadata?: unknown) => 
    logger.performance(component, operation, duration, metadata),
  security: (event: string, details: unknown, component?: string) => logger.security(event, details, component),
  api: (method: string, url: string, status: number, duration: number, error?: unknown) => 
    logger.api(method, url, status, duration, error)
};

export default logger;
