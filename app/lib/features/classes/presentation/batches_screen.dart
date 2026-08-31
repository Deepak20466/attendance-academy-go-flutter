import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/widgets/async_value_widget.dart';
import '../../activities/data/activity_repository.dart';
import '../../coaches/data/coach_repository.dart';
import '../../locations/data/location_repository.dart';
import '../data/class_repository.dart';
import '../data/class_models.dart';

const _dayLabels = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

class BatchesScreen extends ConsumerStatefulWidget {
  const BatchesScreen({super.key});

  @override
  ConsumerState<BatchesScreen> createState() => _BatchesScreenState();
}

class _BatchesScreenState extends ConsumerState<BatchesScreen> {
  String? _activityId;

  @override
  Widget build(BuildContext context) {
    final activities = ref.watch(activitiesListProvider);
    // Resolved up front (not mutated as a side effect deep in `body`) so the
    // FAB — built before `body` in this Scaffold — sees the same value the
    // dropdown defaults to on the very first frame, instead of one build late.
    final loadedActivities = activities.valueOrNull;
    final resolvedActivityId = _activityId ?? (loadedActivities != null && loadedActivities.isNotEmpty ? loadedActivities.first.id : null);

    return Scaffold(
      appBar: AppBar(title: const Text('Batches')),
      floatingActionButton: resolvedActivityId == null
          ? null
          : FloatingActionButton.extended(
              onPressed: () => _showCreateDialog(context, ref, resolvedActivityId),
              icon: const Icon(Icons.add),
              label: const Text('Add batch'),
            ),
      body: AsyncValueWidget(
        value: activities,
        data: (items) {
          if (items.isEmpty) {
            return const Center(
              child: Padding(
                padding: EdgeInsets.all(32),
                child: Text('Create an activity first before adding batches.', style: TextStyle(color: Colors.grey)),
              ),
            );
          }
          return Column(
            children: [
              Padding(
                padding: const EdgeInsets.all(16),
                child: DropdownButtonFormField<String>(
                  initialValue: resolvedActivityId,
                  decoration: const InputDecoration(labelText: 'Activity'),
                  items: items.map((a) => DropdownMenuItem(value: a.id, child: Text(a.name))).toList(),
                  onChanged: (v) => setState(() => _activityId = v),
                ),
              ),
              Expanded(child: _BatchesList(activityId: resolvedActivityId!)),
            ],
          );
        },
      ),
    );
  }

  Future<void> _showCreateDialog(BuildContext context, WidgetRef ref, String activityId) async {
    final coaches = await ref.read(coachRepositoryProvider).list(activityId: activityId, pageSize: 200);
    final locations = await ref.read(locationRepositoryProvider).listByActivity(activityId);
    if (!context.mounted) return;

    final nameController = TextEditingController();
    final descController = TextEditingController();
    TimeOfDay startTime = const TimeOfDay(hour: 16, minute: 0);
    TimeOfDay endTime = const TimeOfDay(hour: 17, minute: 0);
    String? defaultCoachId;
    String? locationId;
    final selectedDays = <int>{};
    String? errorText;

    await showDialog<void>(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setState) => AlertDialog(
          title: const Text('Add batch'),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                TextField(controller: nameController, decoration: const InputDecoration(labelText: 'Name')),
                const SizedBox(height: 12),
                TextField(controller: descController, decoration: const InputDecoration(labelText: 'Description')),
                const SizedBox(height: 12),
                DropdownButtonFormField<String?>(
                  initialValue: defaultCoachId,
                  decoration: const InputDecoration(labelText: 'Default coach (optional)'),
                  items: [
                    const DropdownMenuItem(value: null, child: Text('None')),
                    ...coaches.data.map((c) => DropdownMenuItem(value: c.id, child: Text(c.name))),
                  ],
                  onChanged: (v) => setState(() => defaultCoachId = v),
                ),
                const SizedBox(height: 12),
                DropdownButtonFormField<String?>(
                  initialValue: locationId,
                  decoration: const InputDecoration(labelText: 'Location (optional)'),
                  items: [
                    const DropdownMenuItem(value: null, child: Text('None')),
                    ...locations.map((l) => DropdownMenuItem(value: l.id, child: Text(l.name))),
                  ],
                  onChanged: (v) => setState(() => locationId = v),
                ),
                const SizedBox(height: 12),
                const Align(alignment: Alignment.centerLeft, child: Text('Days of week')),
                const SizedBox(height: 4),
                Wrap(
                  spacing: 6,
                  children: List.generate(7, (i) {
                    return FilterChip(
                      label: Text(_dayLabels[i]),
                      selected: selectedDays.contains(i),
                      onSelected: (sel) => setState(() => sel ? selectedDays.add(i) : selectedDays.remove(i)),
                    );
                  }),
                ),
                const SizedBox(height: 12),
                ListTile(
                  contentPadding: EdgeInsets.zero,
                  title: Text('Start time: ${startTime.format(context)}'),
                  trailing: const Icon(Icons.access_time),
                  onTap: () async {
                    final picked = await showTimePicker(context: context, initialTime: startTime);
                    if (picked != null) setState(() => startTime = picked);
                  },
                ),
                ListTile(
                  contentPadding: EdgeInsets.zero,
                  title: Text('End time: ${endTime.format(context)}'),
                  trailing: const Icon(Icons.access_time),
                  onTap: () async {
                    final picked = await showTimePicker(context: context, initialTime: endTime);
                    if (picked != null) setState(() => endTime = picked);
                  },
                ),
                if (errorText != null) ...[
                  const SizedBox(height: 8),
                  Text(errorText!, style: const TextStyle(color: Colors.red)),
                ],
              ],
            ),
          ),
          actions: [
            TextButton(onPressed: () => Navigator.pop(context), child: const Text('Cancel')),
            FilledButton(
              onPressed: () async {
                if (nameController.text.trim().isEmpty || selectedDays.isEmpty) {
                  setState(() => errorText = 'Name and at least one day are required.');
                  return;
                }
                try {
                  await ref.read(classRepositoryProvider).createBatch(
                        activityId: activityId,
                        name: nameController.text.trim(),
                        description: descController.text.trim(),
                        defaultCoachId: defaultCoachId,
                        locationId: locationId,
                        daysOfWeek: selectedDays.toList()..sort(),
                        startTime: _fmtTime(startTime),
                        endTime: _fmtTime(endTime),
                      );
                  ref.invalidate(_batchesForActivityProvider(activityId));
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

String _fmtTime(TimeOfDay t) => '${t.hour.toString().padLeft(2, '0')}:${t.minute.toString().padLeft(2, '0')}';

class _BatchesList extends ConsumerWidget {
  const _BatchesList({required this.activityId});
  final String activityId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final batches = ref.watch(_batchesForActivityProvider(activityId));
    return RefreshIndicator(
      onRefresh: () => ref.refresh(_batchesForActivityProvider(activityId).future),
      child: AsyncValueWidget(
        value: batches,
        data: (items) {
          if (items.isEmpty) {
            return ListView(
              children: const [
                Padding(
                  padding: EdgeInsets.all(32),
                  child: Center(child: Text('No batches found for this activity.', style: TextStyle(color: Colors.grey))),
                ),
              ],
            );
          }
          return ListView.separated(
            padding: const EdgeInsets.symmetric(vertical: 8),
            itemCount: items.length,
            separatorBuilder: (_, __) => const Divider(height: 1),
            itemBuilder: (context, index) {
              final b = items[index];
              final days = b.daysOfWeek.map((d) => _dayLabels[d]).join(', ');
              return ListTile(
                leading: const Icon(Icons.event_repeat_outlined),
                title: Text(b.name),
                subtitle: Text('$days • ${b.startTime}-${b.endTime}'),
              );
            },
          );
        },
      ),
    );
  }
}

final _batchesForActivityProvider = FutureProvider.autoDispose.family<List<BatchModel>, String>((ref, activityId) {
  return ref.watch(classRepositoryProvider).listBatches(activityId: activityId);
});
