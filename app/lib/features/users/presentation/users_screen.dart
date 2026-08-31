import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../../core/widgets/async_value_widget.dart';
import '../../auth/data/auth_controller.dart';
import '../data/user_repository.dart';

/// Admin-only account management: creating further admin logins and
/// activating/deactivating them. Coach accounts are created from the
/// Coaches screen instead, which also provisions the coach profile and
/// activity assignments — this screen only ever talks to `/users`, which
/// the backend restricts to admin-role accounts.
class UsersScreen extends ConsumerWidget {
  const UsersScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final session = ref.watch(authControllerProvider).valueOrNull;
    final isAdmin = session?.isAdmin ?? false;

    if (!isAdmin) {
      return Scaffold(
        appBar: AppBar(title: const Text('Users')),
        body: const Center(
          child: Padding(
            padding: EdgeInsets.all(32),
            child: Text('This screen is restricted to admins.', style: TextStyle(color: Colors.grey)),
          ),
        ),
      );
    }

    final users = ref.watch(_usersListProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Users')),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _showCreateDialog(context, ref),
        icon: const Icon(Icons.person_add_outlined),
        label: const Text('Add admin'),
      ),
      body: RefreshIndicator(
        onRefresh: () => ref.refresh(_usersListProvider.future),
        child: AsyncValueWidget(
          value: users,
          data: (result) {
            if (result.data.isEmpty) {
              return ListView(
                children: const [
                  Padding(
                    padding: EdgeInsets.all(32),
                    child: Center(child: Text('No users found.', style: TextStyle(color: Colors.grey))),
                  ),
                ],
              );
            }
            return ListView.separated(
              padding: const EdgeInsets.symmetric(vertical: 8),
              itemCount: result.data.length,
              separatorBuilder: (_, __) => const Divider(height: 1),
              itemBuilder: (context, index) {
                final user = result.data[index];
                final isSelf = user.id == session?.userId;
                return ListTile(
                  leading: CircleAvatar(child: Text(user.name.isNotEmpty ? user.name[0].toUpperCase() : '?')),
                  title: Text(user.name),
                  subtitle: Text(
                    '${user.email} • ${user.role} • since ${DateFormat.yMMMd().format(user.createdAt)}',
                  ),
                  trailing: Switch(
                    value: user.isActive,
                    onChanged: isSelf
                        ? null
                        : (v) async {
                            try {
                              await ref.read(userRepositoryProvider).setActive(user.id, v);
                              ref.invalidate(_usersListProvider);
                            } catch (e) {
                              if (context.mounted) {
                                ScaffoldMessenger.of(context).showSnackBar(
                                  SnackBar(content: Text(_errorMessage(e))),
                                );
                              }
                            }
                          },
                  ),
                );
              },
            );
          },
        ),
      ),
    );
  }

  Future<void> _showCreateDialog(BuildContext context, WidgetRef ref) async {
    final nameController = TextEditingController();
    final emailController = TextEditingController();
    final phoneController = TextEditingController();
    final passwordController = TextEditingController();
    String? errorText;

    await showDialog<void>(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setState) => AlertDialog(
          title: const Text('Add admin user'),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                TextField(controller: nameController, decoration: const InputDecoration(labelText: 'Name')),
                const SizedBox(height: 12),
                TextField(
                  controller: emailController,
                  keyboardType: TextInputType.emailAddress,
                  decoration: const InputDecoration(labelText: 'Email'),
                ),
                const SizedBox(height: 12),
                TextField(controller: phoneController, decoration: const InputDecoration(labelText: 'Phone')),
                const SizedBox(height: 12),
                TextField(
                  controller: passwordController,
                  obscureText: true,
                  decoration: const InputDecoration(labelText: 'Password (min 8 characters)'),
                ),
                const SizedBox(height: 12),
                // The backend's /users endpoint only ever provisions admin
                // accounts (coaches go through /coaches instead, which also
                // sets up their coach profile), so role is shown as fixed
                // rather than a selector implying a choice the API doesn't
                // actually offer.
                InputDecorator(
                  decoration: const InputDecoration(labelText: 'Role'),
                  child: const Text('Admin', style: TextStyle(fontWeight: FontWeight.w600)),
                ),
                if (errorText != null) ...[
                  const SizedBox(height: 12),
                  Text(errorText!, style: const TextStyle(color: Colors.red)),
                ],
              ],
            ),
          ),
          actions: [
            TextButton(onPressed: () => Navigator.pop(context), child: const Text('Cancel')),
            FilledButton(
              onPressed: () async {
                final name = nameController.text.trim();
                final email = emailController.text.trim();
                final password = passwordController.text;
                if (name.isEmpty) {
                  setState(() => errorText = 'Name is required.');
                  return;
                }
                if (!_isValidEmail(email)) {
                  setState(() => errorText = 'Enter a valid email address.');
                  return;
                }
                if (password.length < 8) {
                  setState(() => errorText = 'Password must be at least 8 characters.');
                  return;
                }
                try {
                  await ref.read(userRepositoryProvider).createAdmin(
                        name: name,
                        email: email,
                        phone: phoneController.text.trim(),
                        password: password,
                      );
                  ref.invalidate(_usersListProvider);
                  if (context.mounted) {
                    Navigator.pop(context);
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(content: Text('Admin account created for $email.')),
                    );
                  }
                } catch (e) {
                  setState(() => errorText = _errorMessage(e));
                }
              },
              child: const Text('Add'),
            ),
          ],
        ),
      ),
    );
  }
}

bool _isValidEmail(String value) {
  return RegExp(r'^[^@\s]+@[^@\s]+\.[^@\s]+$').hasMatch(value);
}

/// Surfaces the backend's actual error message (e.g. "a user with this
/// email already exists") instead of a raw exception dump, falling back to
/// a generic message for network failures or unexpected responses.
String _errorMessage(Object e) {
  if (e is DioException) {
    final status = e.response?.statusCode;
    final data = e.response?.data;
    if (data is Map && data['error'] is String) {
      return data['error'] as String;
    }
    if (status == 403) return 'You do not have permission to do this.';
    if (status != null) return 'Request failed (HTTP $status).';
    return 'Network error — check your connection and try again.';
  }
  return 'Something went wrong: $e';
}

// Scoped to admin accounts: coach logins already have their own management
// surface on the Coaches screen, and mixing both here would blur what each
// screen is for.
final _usersListProvider = FutureProvider.autoDispose((ref) {
  return ref.watch(userRepositoryProvider).list(role: 'admin');
});
