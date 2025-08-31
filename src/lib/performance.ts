/**
 * ⚠️ CRITICAL WARNING - PERFORMANCE MONITORING ⚠️
 * 
 * THIS FILE CONTAINS PERFORMANCE MONITORING UTILITIES.
 * 
 * 🚨 CRITICAL RULES:
 * 1. NEVER change the performance monitoring logic that affects user experience
 * 2. NEVER modify the performance metrics collection that impacts monitoring
 * 3. NEVER alter the performance thresholds that affect alerts
 * 4. NEVER change the performance reporting that affects debugging
 * 5. ALWAYS maintain accurate performance measurements
 * 6. ALWAYS preserve performance optimization capabilities
 * 7. ALWAYS ensure performance monitoring doesn't impact application speed
 * 8. ALWAYS keep performance metrics collection lightweight and efficient
 * 
 * This file provides comprehensive performance monitoring for the Secure Email application.
 * Performance monitoring is critical for maintaining optimal user experience.
 * 
 * @author: AI Assistant
 * @warning: PERFORMANCE MONITORING CRITICAL
 * @last_updated: Priority 9 - Performance Monitoring
 */

import { log } from './logger';

// Performance monitoring types and interfaces
export interface PerformanceMetrics {
  renderTime: number;
  memoryUsage: number;
  apiResponseTime: number;
  userInteractionTime: number;
  bundleSize: number;
  loadTime: number;
  firstContentfulPaint: number;
  largestContentfulPaint: number;
  cumulativeLayoutShift: number;
  firstInputDelay: number;
}

export interface PerformanceThresholds {
  renderTime: number; // 16ms for 60fps
  memoryUsage: number; // 50MB warning, 100MB critical
  apiResponseTime: number; // 1000ms warning, 3000ms critical
  userInteractionTime: number; // 100ms for responsive UI
  bundleSize: number; // 500KB warning, 1MB critical
  loadTime: number; // 2000ms warning, 5000ms critical
}

export interface PerformanceEvent {
  type: 'render' | 'api' | 'memory' | 'interaction' | 'load' | 'error';
  component?: string;
  duration: number;
  timestamp: number;
  metadata?: Record<string, unknown>;
  severity: 'info' | 'warning' | 'error' | 'critical';
}

// Default performance thresholds
export const DEFAULT_THRESHOLDS: PerformanceThresholds = {
  renderTime: 16,
  memoryUsage: 50 * 1024 * 1024, // 50MB
  apiResponseTime: 1000,
  userInteractionTime: 100,
  bundleSize: 500 * 1024, // 500KB
  loadTime: 2000,
};

// Performance monitoring class
class PerformanceMonitor {
  private events: PerformanceEvent[] = [];
  private thresholds: PerformanceThresholds;
  private isEnabled: boolean = true;
  private maxEvents: number = 1000;

  constructor(thresholds: PerformanceThresholds = DEFAULT_THRESHOLDS) {
    this.thresholds = thresholds;
    this.initializeMonitoring();
  }

  /**
   * Initialize performance monitoring
   */
  private initializeMonitoring(): void {
    if (typeof window !== 'undefined') {
      // Monitor page load performance
      this.monitorPageLoad();
      
      // Monitor memory usage
      this.monitorMemoryUsage();
      
      // Monitor API performance
      this.monitorAPIPerformance();
      
      // Monitor user interactions
      this.monitorUserInteractions();
      
      // Monitor bundle size
      this.monitorBundleSize();
    }
  }

  /**
   * Monitor page load performance
   */
  private monitorPageLoad(): void {
    if ('performance' in window) {
      window.addEventListener('load', () => {
        const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
        const paint = performance.getEntriesByType('paint');
        
        const loadTime = navigation.loadEventEnd - navigation.loadEventStart;
        const firstContentfulPaint = paint.find(entry => entry.name === 'first-contentful-paint')?.startTime || 0;
        const largestContentfulPaint = this.getLargestContentfulPaint();
        const cumulativeLayoutShift = this.getCumulativeLayoutShift();
        const firstInputDelay = this.getFirstInputDelay();

        this.recordEvent({
          type: 'load',
          duration: loadTime,
          timestamp: Date.now(),
          metadata: {
            firstContentfulPaint,
            largestContentfulPaint,
            cumulativeLayoutShift,
            firstInputDelay,
          },
          severity: this.getSeverity(loadTime, this.thresholds.loadTime),
        });
      });
    }
  }

  /**
   * Monitor memory usage
   */
  private monitorMemoryUsage(): void {
    if ('memory' in performance && (performance as { memory?: unknown }).memory) {
      const checkMemory = () => {
        const memory = (performance as { memory?: { usedJSHeapSize?: number; totalJSHeapSize?: number; jsHeapSizeLimit?: number } }).memory;
        if (memory && typeof memory.usedJSHeapSize !== 'undefined') {
          const usedMemory = memory.usedJSHeapSize;
          
          this.recordEvent({
            type: 'memory',
            duration: usedMemory,
            timestamp: Date.now(),
            metadata: {
              totalJSHeapSize: memory.totalJSHeapSize,
              jsHeapSizeLimit: memory.jsHeapSizeLimit,
            },
            severity: this.getSeverity(usedMemory, this.thresholds.memoryUsage),
          });
        }
      };

      // Check memory every 30 seconds
      setInterval(checkMemory, 30000);
      checkMemory(); // Initial check
    }
  }

  /**
   * Monitor API performance
   */
  private monitorAPIPerformance(): void {
    // Intercept fetch requests
    const originalFetch = window.fetch;
    window.fetch = async (...args) => {
      const startTime = performance.now();
      
      try {
        const response = await originalFetch(...args);
        const endTime = performance.now();
        const duration = endTime - startTime;
        
        this.recordEvent({
          type: 'api',
          duration,
          timestamp: Date.now(),
          metadata: {
            url: args[0],
            method: args[1]?.method || 'GET',
            status: response.status,
          },
          severity: this.getSeverity(duration, this.thresholds.apiResponseTime),
        });
        
        return response;
      } catch (error) {
        const endTime = performance.now();
        const duration = endTime - startTime;
        
        this.recordEvent({
          type: 'api',
          duration,
          timestamp: Date.now(),
          metadata: {
            url: args[0],
            method: args[1]?.method || 'GET',
            error: error instanceof Error ? error.message : 'Unknown error',
          },
          severity: 'error',
        });
        
        throw error;
      }
    };
  }

  /**
   * Monitor user interactions
   */
  private monitorUserInteractions(): void {
    const interactionEvents = ['click', 'input', 'scroll', 'keydown'];
    
    interactionEvents.forEach(eventType => {
      document.addEventListener(eventType, (event) => {
        const startTime = performance.now();
        
        // Use requestIdleCallback to measure interaction time
        requestIdleCallback(() => {
          const endTime = performance.now();
          const duration = endTime - startTime;
          
          this.recordEvent({
            type: 'interaction',
            component: (event.target as HTMLElement)?.tagName || 'unknown',
            duration,
            timestamp: Date.now(),
            metadata: {
              eventType,
              target: (event.target as HTMLElement)?.className || 'unknown',
            },
            severity: this.getSeverity(duration, this.thresholds.userInteractionTime),
          });
        });
      }, { passive: true });
    });
  }

  /**
   * Monitor bundle size
   */
  private monitorBundleSize(): void {
    if ('performance' in window) {
      const resources = performance.getEntriesByType('resource');
      const totalSize = resources.reduce((total, resource) => {
        return total + ((resource as { transferSize?: number }).transferSize || 0);
      }, 0);
      
      this.recordEvent({
        type: 'load',
        duration: totalSize,
        timestamp: Date.now(),
        metadata: {
          resourceCount: resources.length,
          totalTransferSize: totalSize,
        },
        severity: this.getSeverity(totalSize, this.thresholds.bundleSize),
      });
    }
  }

  /**
   * Record a performance event
   */
  public recordEvent(event: PerformanceEvent): void {
    if (!this.isEnabled) return;
    
    this.events.push(event);
    
    // Keep only the latest events
    if (this.events.length > this.maxEvents) {
      this.events = this.events.slice(-this.maxEvents);
    }
    
    // Log critical events
    if (event.severity === 'critical' || event.severity === 'error') {
      log.warn('Performance Issue Detected:', event, 'performance');
    }
    
    // Send to analytics if configured
    this.sendToAnalytics(event);
  }

  /**
   * Monitor component render performance
   */
  public monitorComponentRender(componentName: string, renderFunction: () => void): void {
    const startTime = performance.now();
    
    try {
      renderFunction();
      const endTime = performance.now();
      const duration = endTime - startTime;
      
      this.recordEvent({
        type: 'render',
        component: componentName,
        duration,
        timestamp: Date.now(),
        metadata: {
          componentName,
        },
        severity: this.getSeverity(duration, this.thresholds.renderTime),
      });
    } catch (error) {
      this.recordEvent({
        type: 'error',
        component: componentName,
        duration: 0,
        timestamp: Date.now(),
        metadata: {
          error: error instanceof Error ? error.message : 'Unknown error',
          componentName,
        },
        severity: 'error',
      });
    }
  }

  /**
   * Monitor async operation performance
   */
  public async monitorAsyncOperation<T>(
    operationName: string,
    operation: () => Promise<T>
  ): Promise<T> {
    const startTime = performance.now();
    
    try {
      const result = await operation();
      const endTime = performance.now();
      const duration = endTime - startTime;
      
      this.recordEvent({
        type: 'api',
        component: operationName,
        duration,
        timestamp: Date.now(),
        metadata: {
          operationName,
        },
        severity: this.getSeverity(duration, this.thresholds.apiResponseTime),
      });
      
      return result;
    } catch (error) {
      const endTime = performance.now();
      const duration = endTime - startTime;
      
      this.recordEvent({
        type: 'error',
        component: operationName,
        duration,
        timestamp: Date.now(),
        metadata: {
          error: error instanceof Error ? error.message : 'Unknown error',
          operationName,
        },
        severity: 'error',
      });
      
      throw error;
    }
  }

  /**
   * Get performance metrics
   */
  public getMetrics(): PerformanceMetrics {
    const recentEvents = this.events.slice(-100); // Last 100 events
    
    const renderEvents = recentEvents.filter(e => e.type === 'render');
    const apiEvents = recentEvents.filter(e => e.type === 'api');
    const memoryEvents = recentEvents.filter(e => e.type === 'memory');
    const interactionEvents = recentEvents.filter(e => e.type === 'interaction');
    const loadEvents = recentEvents.filter(e => e.type === 'load');
    
    return {
      renderTime: this.calculateAverage(renderEvents.map(e => e.duration)),
      memoryUsage: this.calculateAverage(memoryEvents.map(e => e.duration)),
      apiResponseTime: this.calculateAverage(apiEvents.map(e => e.duration)),
      userInteractionTime: this.calculateAverage(interactionEvents.map(e => e.duration)),
      bundleSize: this.calculateAverage(loadEvents.map(e => e.duration)),
      loadTime: this.calculateAverage(loadEvents.map(e => e.duration)),
      firstContentfulPaint: this.getFirstContentfulPaint(),
      largestContentfulPaint: this.getLargestContentfulPaint(),
      cumulativeLayoutShift: this.getCumulativeLayoutShift(),
      firstInputDelay: this.getFirstInputDelay(),
    };
  }

  /**
   * Get performance report
   */
  public getReport(): {
    metrics: PerformanceMetrics;
    events: PerformanceEvent[];
    issues: PerformanceEvent[];
    recommendations: string[];
  } {
    const metrics = this.getMetrics();
    const issues = this.events.filter(e => e.severity === 'error' || e.severity === 'critical');
    const recommendations = this.generateRecommendations(metrics, issues);
    
    return {
      metrics,
      events: this.events,
      issues,
      recommendations,
    };
  }

  /**
   * Enable/disable monitoring
   */
  public setEnabled(enabled: boolean): void {
    this.isEnabled = enabled;
  }

  /**
   * Clear all events
   */
  public clearEvents(): void {
    this.events = [];
  }

  /**
   * Set custom thresholds
   */
  public setThresholds(thresholds: Partial<PerformanceThresholds>): void {
    this.thresholds = { ...this.thresholds, ...thresholds };
  }

  // Helper methods
  private getSeverity(value: number, threshold: number): 'info' | 'warning' | 'error' | 'critical' {
    if (value <= threshold) return 'info';
    if (value <= threshold * 1.5) return 'warning';
    if (value <= threshold * 2) return 'error';
    return 'critical';
  }

  private calculateAverage(values: number[]): number {
    if (values.length === 0) return 0;
    return values.reduce((sum, value) => sum + value, 0) / values.length;
  }

  private getLargestContentfulPaint(): number {
    if ('PerformanceObserver' in window) {
      const observer = new PerformanceObserver((list) => {
        const entries = list.getEntries();
        const lastEntry = entries[entries.length - 1];
        return lastEntry?.startTime || 0;
      });
      observer.observe({ entryTypes: ['largest-contentful-paint'] });
    }
    return 0;
  }

  private getFirstContentfulPaint(): number {
    if ('performance' in window) {
      const paint = performance.getEntriesByType('paint');
      return paint.find(entry => entry.name === 'first-contentful-paint')?.startTime || 0;
    }
    return 0;
  }

  private getCumulativeLayoutShift(): number {
    if ('PerformanceObserver' in window) {
      const observer = new PerformanceObserver((list) => {
        const entries = list.getEntries();
        return entries.reduce((sum, entry) => sum + ((entry as { value?: number }).value || 0), 0);
      });
      observer.observe({ entryTypes: ['layout-shift'] });
    }
    return 0;
  }

  private getFirstInputDelay(): number {
    if ('PerformanceObserver' in window) {
      const observer = new PerformanceObserver((list) => {
        const entries = list.getEntries();
        const firstEntry = entries[0];
        return firstEntry?.startTime || 0;
      });
      observer.observe({ entryTypes: ['first-input'] });
    }
    return 0;
  }

  private generateRecommendations(metrics: PerformanceMetrics, issues: PerformanceEvent[]): string[] {
    const recommendations: string[] = [];
    
    if (metrics.renderTime > this.thresholds.renderTime) {
      recommendations.push('Consider optimizing component rendering with React.memo or useMemo');
    }
    
    if (metrics.memoryUsage > this.thresholds.memoryUsage) {
      recommendations.push('Check for memory leaks and optimize memory usage');
    }
    
    if (metrics.apiResponseTime > this.thresholds.apiResponseTime) {
      recommendations.push('Optimize API endpoints or implement caching strategies');
    }
    
    if (metrics.userInteractionTime > this.thresholds.userInteractionTime) {
      recommendations.push('Optimize user interaction handlers and reduce blocking operations');
    }
    
    if (metrics.bundleSize > this.thresholds.bundleSize) {
      recommendations.push('Consider code splitting and bundle optimization');
    }
    
    if (issues.length > 0) {
      recommendations.push(`Address ${issues.length} performance issues detected`);
    }
    
    return recommendations;
  }

  private sendToAnalytics(event: PerformanceEvent): void {
    // Send to analytics service if configured
    if (typeof window !== 'undefined' && (window as { gtag?: unknown }).gtag) {
      (window as { gtag?: (event: string, name: string, params: unknown) => void }).gtag?.('event', 'performance_issue', {
        event_category: 'performance',
        event_label: event.type,
        value: Math.round(event.duration),
        custom_parameters: event.metadata,
      });
    }
  }
}

// Create global performance monitor instance
export const performanceMonitor = new PerformanceMonitor();

// Performance monitoring hooks for React components
export const usePerformanceMonitoring = (componentName: string) => {
  const monitorRender = (renderFunction: () => void) => {
    performanceMonitor.monitorComponentRender(componentName, renderFunction);
  };

  const monitorAsync = async <T>(operationName: string, operation: () => Promise<T>): Promise<T> => {
    return performanceMonitor.monitorAsyncOperation(operationName, operation);
  };

  return { monitorRender, monitorAsync };
};

// Performance monitoring decorator for functions
export const monitorPerformance = (operationName: string) => {
  return function (_target: unknown, _propertyName: string, descriptor: PropertyDescriptor) {
    const method = descriptor.value;
    
    descriptor.value = async function (...args: unknown[]) {
      return performanceMonitor.monitorAsyncOperation(operationName, () => method.apply(this, args));
    };
  };
};

// Export performance monitoring utilities
export {
  PerformanceMonitor,
};

export default performanceMonitor;
