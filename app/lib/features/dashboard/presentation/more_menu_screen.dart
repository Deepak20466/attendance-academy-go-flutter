import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../auth/data/auth_controller.dart';

class MoreMenuScreen extends ConsumerWidget {
  const MoreMenuScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final session = ref.watch(authControllerProvider).valueOrNull;
    final isAdmin = session?.isAdmin ?? false;

    return Scaffold(
      appBar: AppBar(title: const Text('More')),
      body: ListView(
        children: [
          ListTile(
            leading: const Icon(Icons.event_busy_outlined),
            title: const Text('Leaves'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push('/leaves'),
          ),
          ListTile(
            leading: const Icon(Icons.payments_outlined),
            title: const Text('Fees'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push('/fees'),
          ),
          ListTile(
            leading: const Icon(Icons.account_balance_wallet_outlined),
            title: const Text('Salary'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push('/salary'),
          ),
          ListTile(
            leading: const Icon(Icons.notifications_outlined),
            title: const Text('Notifications'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push('/notifications'),
          ),
          ListTile(
            leading: const Icon(Icons.swap_horiz_outlined),
            title: Text(isAdmin ? 'Substitutions' : 'My substitute classes'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push('/substitutions'),
          ),
          ListTile(
            leading: const Icon(Icons.location_on_outlined),
            title: Text(isAdmin ? 'Coach Attendance' : 'My Attendance'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push('/coach-attendance'),
          ),
          if (isAdmin)
            ListTile(
              leading: const Icon(Icons.bar_chart_outlined),
              title: const Text('Analytics'),
              trailing: const Icon(Icons.chevron_right),
              onTap: () => context.push('/analytics'),
            ),
          if (isAdmin) ...[
            const Divider(height: 32),
            ListTile(
              leading: const Icon(Icons.sports_outlined),
              title: const Text('Activities'),
              trailing: const Icon(Icons.chevron_right),
              onTap: () => context.push('/activities'),
            ),
            ListTile(
              leading: const Icon(Icons.event_repeat_outlined),
              title: const Text('Batches'),
              trailing: const Icon(Icons.chevron_right),
              onTap: () => context.push('/batches'),
            ),
            ListTile(
              leading: const Icon(Icons.pin_drop_outlined),
              title: const Text('Locations'),
              trailing: const Icon(Icons.chevron_right),
              onTap: () => context.push('/locations'),
            ),
            ListTile(
              leading: const Icon(Icons.admin_panel_settings_outlined),
              title: const Text('Users'),
              trailing: const Icon(Icons.chevron_right),
              onTap: () => context.push('/users'),
            ),
          ],
          const Divider(height: 32),
          ListTile(
            leading: const Icon(Icons.logout, color: Colors.red),
            title: const Text('Log out', style: TextStyle(color: Colors.red)),
            onTap: () async {
              await ref.read(authControllerProvider.notifier).logout();
            },
          ),
        ],
      ),
    );
  }
}
