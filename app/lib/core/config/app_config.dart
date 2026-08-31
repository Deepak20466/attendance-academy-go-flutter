/// Base URL for the Go backend. Override at build time with:
///   flutter run --dart-define=API_BASE_URL=http://10.0.2.2:8080/api/v1
class AppConfig {
  static const String apiBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://10.0.2.2:8080/api/v1',
  );
}
