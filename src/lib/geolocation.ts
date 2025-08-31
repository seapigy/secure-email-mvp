// Mock geolocation functions for testing
export const validateCountryCode = (countryCode: string): boolean => {
  if (!countryCode || typeof countryCode !== 'string') return false;
  const trimmed = countryCode.trim();
  if (trimmed.length !== 2) return false;
  const countryRegex = /^[A-Z]{2}$/;
  return countryRegex.test(trimmed);
};

export const validateCityName = (cityName: string): boolean => {
  if (!cityName || typeof cityName !== 'string') return false;
  const trimmed = cityName.trim();
  if (trimmed.length < 2 || trimmed.length > 100) return false;
  const cityRegex = /^[a-zA-Z\s\-']+$/;
  return cityRegex.test(trimmed);
};

export const SUPPORTED_COUNTRIES = [
  { code: 'US', name: 'United States' },
  { code: 'CA', name: 'Canada' },
  { code: 'GB', name: 'United Kingdom' },
  { code: 'DE', name: 'Germany' },
  { code: 'FR', name: 'France' },
  { code: 'JP', name: 'Japan' },
  { code: 'AU', name: 'Australia' },
  { code: 'BR', name: 'Brazil' },
  { code: 'IN', name: 'India' },
  { code: 'CN', name: 'China' },
];

export enum GeoVerificationType {
  NONE = 'none',
  COUNTRY = 'country',
  CITY = 'city',
  CITY_COUNTRY = 'city_country',
}

// Mock implementations for testing
export const mockGeolocationFunctions = {
  validateCountryCode: (countryCode: string): boolean => {
    return SUPPORTED_COUNTRIES.some(country => country.code === countryCode);
  },
  validateCityName: (cityName: string): boolean => {
    return cityName.length >= 2 && cityName.length <= 100;
  },
  SUPPORTED_COUNTRIES,
  GeoVerificationType,
};



