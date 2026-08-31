import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/widgets/async_value_widget.dart';
import '../data/activity_repository.dart';

class ActivitiesScreen extends ConsumerWidget {
  const ActivitiesScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final activities = ref.watch(_allActivitiesProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Activities')),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _showCreateDialog(context, ref),
        icon: const Icon(Icons.add),
        label: const Text('Add activity'),
      ),
      body: RefreshIndicator(
        onRefresh: () => ref.refresh(_allActivitiesProvider.future),
        child: AsyncValueWidget(
          value: activities,
          data: (items) {
            if (items.isEmpty) {
              return ListView(
                children: const [
                  Padding(
                    padding: EdgeInsets.all(32),
                    child: Center(child: Text('No activities found.', style: TextStyle(color: Colors.grey))),
                  ),
                ],
              );
            }
            return ListView.separated(
              padding: const EdgeInsets.symmetric(vertical: 8),
              itemCount: items.length,
              separatorBuilder: (_, __) => const Divider(height: 1),
              itemBuilder: (context, index) {
                final a = items[index];
                return ListTile(
                  leading: const Icon(Icons.sports_outlined),
                  title: Text(a.name),
                  subtitle: a.description.isNotEmpty ? Text(a.description) : null,
                  trailing: a.isActive ? null : const Chip(label: Text('Inactive')),
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
    final descController = TextEditingController();
    String? errorText;

    await showDialog<void>(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setState) => AlertDialog(
          title: const Text('Add activity'),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              TextField(controller: nameController, decoration: const InputDecoration(labelText: 'Name')),
              const SizedBox(height: 12),
              TextField(controller: descController, decoration: const InputDecoration(labelText: 'Description')),
              if (errorText != null) ...[
                const SizedBox(height: 12),
                Text(errorText!, style: const TextStyle(color: Colors.red)),
              ],
            ],
          ),
          actions: [
            TextButton(onPressed: () => Navigator.pop(context), child: const Text('Cancel')),
            FilledButton(
              onPressed: () async {
                if (nameController.text.trim().isEmpty) {
                  setState(() => errorText = 'Name is required.');
                  return;
                }
                try {
                  await ref.read(activityRepositoryProvider).create(
                        name: nameController.text.trim(),
                        description: descController.text.trim(),
                      );
                  ref.invalidate(_allActivitiesProvider);
                  ref.invalidate(activitiesListProvider);
                  if (context.mounted) Navigator.pop(context);
                } catch (e) {
                  setState(() => errorText = 'Failed: $e');
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

final _allActivitiesProvider = FutureProvider.autoDispose((ref) {
  return ref.watch(activityRepositoryProvider).list(onlyActive: false);
});
