// =============================================================================
// SECURE EMAIL MVP - GEOLOCATION UTILITIES
// =============================================================================
// Frontend utilities for geolocation validation and country/city handling.
// =============================================================================

/**
 * Validates a country code against ISO 3166-1 alpha-2 standard
 * @param countryCode - The country code to validate
 * @returns true if valid, false otherwise
 */
export const validateCountryCode = (countryCode: string): boolean => {
  if (!countryCode || typeof countryCode !== 'string') {
    return false;
  }
  
  const trimmed = countryCode.trim();
  if (trimmed.length !== 2) {
    return false;
  }
  
  // Check for valid format (2 uppercase letters)
  const countryRegex = /^[A-Z]{2}$/;
  return countryRegex.test(trimmed);
};

/**
 * Validates a city name format
 * @param cityName - The city name to validate
 * @returns true if valid, false otherwise
 */
export const validateCityName = (cityName: string): boolean => {
  if (!cityName || typeof cityName !== 'string') {
    return false;
  }
  
  const trimmed = cityName.trim();
  if (trimmed.length < 2 || trimmed.length > 100) {
    return false;
  }
  
  // Check for valid characters (letters, spaces, hyphens, apostrophes)
  const cityRegex = /^[a-zA-Z\s\-']+$/;
  return cityRegex.test(trimmed);
};

/**
 * Normalizes a city name for consistent comparison
 * @param cityName - The city name to normalize
 * @returns Normalized city name
 */
export const normalizeCityName = (cityName: string): string => {
  if (!cityName) {
    return '';
  }
  
  return cityName
    .trim()
    .toLowerCase()
    .replace(/\s+/g, ' ') // Replace multiple spaces with single space
    .replace(/[^\w\s\-']/g, ''); // Remove invalid characters
};

/**
 * Normalizes a country code for consistent comparison
 * @param countryCode - The country code to normalize
 * @returns Normalized country code (uppercase)
 */
export const normalizeCountryCode = (countryCode: string): string => {
  if (!countryCode) {
    return '';
  }
  
  return countryCode.trim().toUpperCase();
};

/**
 * List of supported countries with their ISO codes and names
 */
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
  { code: 'MX', name: 'Mexico' },
  { code: 'IT', name: 'Italy' },
  { code: 'ES', name: 'Spain' },
  { code: 'NL', name: 'Netherlands' },
  { code: 'SE', name: 'Sweden' },
  { code: 'NO', name: 'Norway' },
  { code: 'DK', name: 'Denmark' },
  { code: 'FI', name: 'Finland' },
  { code: 'CH', name: 'Switzerland' },
  { code: 'AT', name: 'Austria' },
  { code: 'BE', name: 'Belgium' },
  { code: 'IE', name: 'Ireland' },
  { code: 'NZ', name: 'New Zealand' },
  { code: 'SG', name: 'Singapore' },
  { code: 'KR', name: 'South Korea' },
  { code: 'IL', name: 'Israel' },
  { code: 'ZA', name: 'South Africa' },
  { code: 'AR', name: 'Argentina' },
  { code: 'CL', name: 'Chile' },
  { code: 'CO', name: 'Colombia' },
  { code: 'PE', name: 'Peru' },
  { code: 'VE', name: 'Venezuela' },
  { code: 'MY', name: 'Malaysia' },
  { code: 'TH', name: 'Thailand' },
  { code: 'VN', name: 'Vietnam' },
  { code: 'PH', name: 'Philippines' },
  { code: 'ID', name: 'Indonesia' },
  { code: 'TR', name: 'Turkey' },
  { code: 'PL', name: 'Poland' },
  { code: 'CZ', name: 'Czech Republic' },
  { code: 'HU', name: 'Hungary' },
  { code: 'RO', name: 'Romania' },
  { code: 'BG', name: 'Bulgaria' },
  { code: 'HR', name: 'Croatia' },
  { code: 'SI', name: 'Slovenia' },
  { code: 'SK', name: 'Slovakia' },
  { code: 'LT', name: 'Lithuania' },
  { code: 'LV', name: 'Latvia' },
  { code: 'EE', name: 'Estonia' },
  { code: 'LU', name: 'Luxembourg' },
  { code: 'MT', name: 'Malta' },
  { code: 'CY', name: 'Cyprus' },
  { code: 'GR', name: 'Greece' },
  { code: 'PT', name: 'Portugal' },
] as const;

/**
 * Gets a country name by its ISO code
 * @param countryCode - The ISO country code
 * @returns The country name or undefined if not found
 */
export const getCountryName = (countryCode: string): string | undefined => {
  const country = SUPPORTED_COUNTRIES.find(c => c.code === countryCode.toUpperCase());
  return country?.name;
};

/**
 * Gets a country code by its name
 * @param countryName - The country name
 * @returns The ISO country code or undefined if not found
 */
export const getCountryCode = (countryName: string): string | undefined => {
  const country = SUPPORTED_COUNTRIES.find(c => 
    c.name.toLowerCase() === countryName.toLowerCase()
  );
  return country?.code;
};

/**
 * Geolocation verification types
 */
export type GeoVerificationType = 'none' | 'country' | 'city' | 'city_country';

/**
 * Validates a geolocation verification type
 * @param verificationType - The verification type to validate
 * @returns true if valid, false otherwise
 */
export const validateGeoVerificationType = (verificationType: string): verificationType is GeoVerificationType => {
  return ['none', 'country', 'city', 'city_country'].includes(verificationType);
};

/**
 * Gets the description for a geolocation verification type
 * @param verificationType - The verification type
 * @returns Description of the verification type
 */
export const getGeoVerificationDescription = (verificationType: GeoVerificationType): string => {
  switch (verificationType) {
    case 'none':
      return 'No location restrictions';
    case 'country':
      return 'Restrict access by country only';
    case 'city':
      return 'Restrict access by city only';
    case 'city_country':
      return 'Restrict access by both city and country';
    default:
      return 'Unknown verification type';
  }
};
