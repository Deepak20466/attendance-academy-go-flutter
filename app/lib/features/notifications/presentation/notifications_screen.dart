import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../../core/widgets/async_value_widget.dart';
import '../data/notification_repository.dart';

class NotificationsScreen extends ConsumerWidget {
  const NotificationsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final notifications = ref.watch(notificationsListProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Notifications')),
      body: RefreshIndicator(
        onRefresh: () => ref.refresh(notificationsListProvider.future),
        child: AsyncValueWidget(
          value: notifications,
          data: (result) {
            if (result.data.isEmpty) {
              return ListView(
                children: const [
                  Padding(
                    padding: EdgeInsets.all(32),
                    child: Center(child: Text('No notifications yet.', style: TextStyle(color: Colors.grey))),
                  ),
                ],
              );
            }
            return ListView.separated(
              itemCount: result.data.length,
              separatorBuilder: (_, __) => const Divider(height: 1),
              itemBuilder: (context, index) {
                final n = result.data[index];
                return ListTile(
                  leading: Icon(
                    n.isRead ? Icons.notifications_none : Icons.notifications_active,
                    color: n.isRead ? Colors.grey : Theme.of(context).colorScheme.primary,
                  ),
                  title: Text(n.title, style: TextStyle(fontWeight: n.isRead ? FontWeight.normal : FontWeight.w700)),
                  subtitle: Text(n.body),
                  trailing: Text(DateFormat.MMMd().format(n.createdAt), style: const TextStyle(color: Colors.grey, fontSize: 12)),
                  onTap: () async {
                    if (!n.isRead) {
                      await ref.read(notificationRepositoryProvider).markRead(n.id);
                      ref.invalidate(notificationsListProvider);
                    }
                  },
                );
              },
            );
          },
        ),
      ),
    );
  }
}
