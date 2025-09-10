// Authentication Navigator for SecureMail
import React from 'react';
import { createStackNavigator } from '@react-navigation/stack';
import { AuthStackParamList } from '../types';

// Screens
import WebsiteSignupScreen from '../screens/WebsiteSignupScreen';
import WebsiteLoginScreen from '../screens/WebsiteLoginScreen';
import ForgotPasswordScreen from '../screens/ForgotPasswordScreen';
import PlanSelectionScreen from '../screens/PlanSelectionScreen';
import EmailVerificationScreen from '../screens/EmailVerificationScreen';
import RecoveryKeyScreen from '../screens/RecoveryKeyScreen';
import AccountRecoveryScreen from '../screens/AccountRecoveryScreen';

const Stack = createStackNavigator<AuthStackParamList>();

export default function AuthNavigator() {
  return (
    <Stack.Navigator
      initialRouteName="Login"
      screenOptions={{
        headerShown: false,
        cardStyle: { backgroundColor: '#1B1F23' }, // Dark background matching website
      }}
    >
      <Stack.Screen name="Login" component={WebsiteLoginScreen} />
      <Stack.Screen name="PlanSelection" component={PlanSelectionScreen} />
      <Stack.Screen name="Signup" component={WebsiteSignupScreen} />
      <Stack.Screen name="EmailVerification" component={EmailVerificationScreen} />
      <Stack.Screen name="RecoveryKey" component={RecoveryKeyScreen} />
      <Stack.Screen name="AccountRecovery" component={AccountRecoveryScreen} />
      <Stack.Screen name="ForgotPassword" component={ForgotPasswordScreen} />
    </Stack.Navigator>
  );
}
