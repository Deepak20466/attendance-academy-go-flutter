import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/providers/core_providers.dart';
import 'auth_models.dart';
import 'auth_repository.dart';

final authRepositoryProvider = Provider<AuthRepository>((ref) {
  final apiClient = ref.watch(apiClientProvider);
  return AuthRepository(apiClient.dio, ref.watch(tokenStorageProvider));
});

/// AsyncValue<AuthSession?> semantics: loading = still restoring session on
/// app start, data(null) = signed out, data(session) = signed in.
class AuthController extends AsyncNotifier<AuthSession?> {
  @override
  Future<AuthSession?> build() async {
    final apiClient = ref.watch(apiClientProvider);
    apiClient.loggedOutStream.listen((_) {
      state = const AsyncData(null);
    });
    return ref.read(authRepositoryProvider).restoreSession();
  }

  Future<void> login(String email, String password) async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(() => ref.read(authRepositoryProvider).login(email, password));
  }

  Future<void> logout() async {
    await ref.read(authRepositoryProvider).logout();
    state = const AsyncData(null);
  }
}

final authControllerProvider = AsyncNotifierProvider<AuthController, AuthSession?>(AuthController.new);
