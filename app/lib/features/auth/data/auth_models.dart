class AuthSession {
  AuthSession({
    required this.userId,
    required this.name,
    required this.email,
    required this.role,
  });

  factory AuthSession.fromLoginJson(Map<String, dynamic> json) => AuthSession(
        userId: json['user_id'] as String,
        name: json['name'] as String,
        email: json['email'] as String,
        role: json['role'] as String,
      );

  final String userId;
  final String name;
  final String email;
  final String role; // "admin" or "coach"

  bool get isAdmin => role == 'admin';
  bool get isCoach => role == 'coach';
}
