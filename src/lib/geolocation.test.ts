import { 
  validateCountryCode, 
  validateCityName, 
  normalizeCityName, 
  normalizeCountryCode,
  getCountryName,
  getCountryCode,
  validateGeoVerificationType,
  getGeoVerificationDescription,
  SUPPORTED_COUNTRIES
} from './geolocation';

describe('Geolocation Utilities', () => {
  describe('validateCountryCode', () => {
    it('should validate correct country codes', () => {
      expect(validateCountryCode('US')).toBe(true);
      expect(validateCountryCode('CA')).toBe(true);
      expect(validateCountryCode('GB')).toBe(true);
      expect(validateCountryCode('DE')).toBe(true);
    });

    it('should reject invalid country codes', () => {
      expect(validateCountryCode('')).toBe(false);
      expect(validateCountryCode('USA')).toBe(false);
      expect(validateCountryCode('us')).toBe(false);
      expect(validateCountryCode('12')).toBe(false);
      expect(validateCountryCode('A1')).toBe(false);
      expect(validateCountryCode('A')).toBe(false);
    });
  });

  describe('validateCityName', () => {
    it('should validate correct city names', () => {
      expect(validateCityName('New York')).toBe(true);
      expect(validateCityName('London')).toBe(true);
      expect(validateCityName('San Francisco')).toBe(true);
      expect(validateCityName('New York')).toBe(true);
      expect(validateCityName('Saint-Denis')).toBe(true);
      expect(validateCityName("O'Connor")).toBe(true);
    });

    it('should reject invalid city names', () => {
      expect(validateCityName('')).toBe(false);
      expect(validateCityName('A')).toBe(false);
      expect(validateCityName('New York123')).toBe(false);
      expect(validateCityName('New York!')).toBe(false);
      expect(validateCityName('New York@')).toBe(false);
      expect(validateCityName('a'.repeat(101))).toBe(false);
    });
  });

  describe('normalizeCityName', () => {
    it('should normalize city names correctly', () => {
      expect(normalizeCityName('New York')).toBe('new york');
      expect(normalizeCityName('  San Francisco  ')).toBe('san francisco');
      expect(normalizeCityName('New   York')).toBe('new york');
      expect(normalizeCityName('')).toBe('');
    });
  });

  describe('normalizeCountryCode', () => {
    it('should normalize country codes correctly', () => {
      expect(normalizeCountryCode('us')).toBe('US');
      expect(normalizeCountryCode('  CA  ')).toBe('CA');
      expect(normalizeCountryCode('gb')).toBe('GB');
    });
  });

  describe('getCountryName', () => {
    it('should return country name for valid codes', () => {
      expect(getCountryName('US')).toBe('United States');
      expect(getCountryName('CA')).toBe('Canada');
      expect(getCountryName('GB')).toBe('United Kingdom');
    });

    it('should return undefined for invalid codes', () => {
      expect(getCountryName('XX')).toBeUndefined();
      expect(getCountryName('')).toBeUndefined();
    });
  });

  describe('getCountryCode', () => {
    it('should return country code for valid names', () => {
      expect(getCountryCode('United States')).toBe('US');
      expect(getCountryCode('Canada')).toBe('CA');
      expect(getCountryCode('United Kingdom')).toBe('GB');
    });

    it('should return undefined for invalid names', () => {
      expect(getCountryCode('Invalid Country')).toBeUndefined();
      expect(getCountryCode('')).toBeUndefined();
    });
  });

  describe('validateGeoVerificationType', () => {
    it('should validate correct verification types', () => {
      expect(validateGeoVerificationType('none')).toBe(true);
      expect(validateGeoVerificationType('country')).toBe(true);
      expect(validateGeoVerificationType('city')).toBe(true);
      expect(validateGeoVerificationType('city_country')).toBe(true);
    });

    it('should reject invalid verification types', () => {
      expect(validateGeoVerificationType('invalid')).toBe(false);
      expect(validateGeoVerificationType('')).toBe(false);
      expect(validateGeoVerificationType('CITY')).toBe(false);
    });
  });

  describe('getGeoVerificationDescription', () => {
    it('should return correct descriptions', () => {
      expect(getGeoVerificationDescription('none')).toBe('No location restrictions');
      expect(getGeoVerificationDescription('country')).toBe('Restrict access by country only');
      expect(getGeoVerificationDescription('city')).toBe('Restrict access by city only');
      expect(getGeoVerificationDescription('city_country')).toBe('Restrict access by both city and country');
    });
  });

  describe('SUPPORTED_COUNTRIES', () => {
    it('should contain valid country data', () => {
      expect(SUPPORTED_COUNTRIES.length).toBeGreaterThan(0);
      
      SUPPORTED_COUNTRIES.forEach(country => {
        expect(country.code).toMatch(/^[A-Z]{2}$/);
        expect(country.name).toBeTruthy();
        expect(typeof country.name).toBe('string');
      });
    });

    it('should have unique country codes', () => {
      const codes = SUPPORTED_COUNTRIES.map(c => c.code);
      const uniqueCodes = new Set(codes);
      expect(uniqueCodes.size).toBe(codes.length);
    });
  });
});
