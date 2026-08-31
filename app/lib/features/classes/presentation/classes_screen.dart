import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';

import '../../../core/widgets/async_value_widget.dart';
import '../../activities/data/activity_repository.dart';
import '../../auth/data/auth_controller.dart';
import '../../coaches/data/coach_repository.dart';
import '../data/class_models.dart';
import '../data/class_repository.dart';

class ClassesScreen extends ConsumerWidget {
  const ClassesScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final classes = ref.watch(recentClassesProvider);
    final session = ref.watch(authControllerProvider).valueOrNull;
    final isCoach = session?.isCoach ?? false;
    final isAdmin = session?.isAdmin ?? false;

    return Scaffold(
      appBar: AppBar(title: const Text('Classes')),
      floatingActionButton: isAdmin
          ? FloatingActionButton.extended(
              onPressed: () => _showCreateClassDialog(context, ref),
              icon: const Icon(Icons.add),
              label: const Text('Add class'),
            )
          : null,
      body: RefreshIndicator(
        onRefresh: () => ref.refresh(recentClassesProvider.future),
        child: AsyncValueWidget(
          value: classes,
          data: (result) {
            if (result.data.isEmpty) {
              return ListView(
                children: const [
                  Padding(
                    padding: EdgeInsets.all(32),
                    child: Center(child: Text('No classes found.', style: TextStyle(color: Colors.grey))),
                  ),
                ],
              );
            }
            return ListView.separated(
              padding: const EdgeInsets.symmetric(vertical: 8),
              itemCount: result.data.length,
              separatorBuilder: (_, __) => const Divider(height: 1),
              itemBuilder: (context, index) {
                final c = result.data[index];
                return ListTile(
                  leading: Icon(
                    c.attendanceMarked ? Icons.check_circle_outline : Icons.schedule_outlined,
                    color: c.attendanceMarked ? Colors.green : Colors.orange,
                  ),
                  title: Text('${DateFormat.yMMMd().format(c.classDate)} • ${c.startTime}-${c.endTime}'),
                  subtitle: Text(c.attendanceMarked ? 'Attendance submitted' : 'Attendance pending'),
                  trailing: isCoach
                      ? IconButton(
                          icon: const Icon(Icons.location_on_outlined),
                          onPressed: () async {
                            await context.push('/coach-checkin/${c.id}');
                            ref.invalidate(recentClassesProvider);
                          },
                        )
                      : null,
                  onTap: () async {
                    await context.push('/classes/${c.id}/attendance');
                    // The pushed screen doesn't know about this list, so
                    // refresh here on return rather than relying on it to
                    // invalidate a provider it has no reason to know about.
                    ref.invalidate(recentClassesProvider);
                  },
                );
              },
            );
          },
        ),
      ),
    );
  }

  Future<void> _showCreateClassDialog(BuildContext context, WidgetRef ref) async {
    final activities = await ref.read(activityRepositoryProvider).list();
    if (!context.mounted) return;
    if (activities.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Create an activity first.')),
      );
      return;
    }

    String activityId = activities.first.id;
    List<BatchModel> batches = await ref.read(classRepositoryProvider).listBatches(activityId: activityId);
    var coaches = await ref.read(coachRepositoryProvider).list(activityId: activityId, pageSize: 200);
    if (!context.mounted) return;

    String? batchId = batches.isNotEmpty ? batches.first.id : null;
    String? coachId = coaches.data.isNotEmpty ? coaches.data.first.id : null;
    DateTime classDate = DateTime.now();
    TimeOfDay startTime = const TimeOfDay(hour: 16, minute: 0);
    TimeOfDay endTime = const TimeOfDay(hour: 17, minute: 0);
    String? errorText;

    await showDialog<void>(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setState) => AlertDialog(
          title: const Text('Add class'),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                DropdownButtonFormField<String>(
                  initialValue: activityId,
                  decoration: const InputDecoration(labelText: 'Activity'),
                  items: activities.map((a) => DropdownMenuItem(value: a.id, child: Text(a.name))).toList(),
                  onChanged: (v) async {
                    if (v == null) return;
                    final newBatches = await ref.read(classRepositoryProvider).listBatches(activityId: v);
                    final newCoaches = await ref.read(coachRepositoryProvider).list(activityId: v, pageSize: 200);
                    setState(() {
                      activityId = v;
                      batches = newBatches;
                      coaches = newCoaches;
                      batchId = newBatches.isNotEmpty ? newBatches.first.id : null;
                      coachId = newCoaches.data.isNotEmpty ? newCoaches.data.first.id : null;
                    });
                  },
                ),
                const SizedBox(height: 12),
                if (batches.isEmpty)
                  const Text('No batches for this activity — create a batch first.', style: TextStyle(color: Colors.orange))
                else
                  DropdownButtonFormField<String>(
                    initialValue: batchId,
                    decoration: const InputDecoration(labelText: 'Batch'),
                    items: batches.map((b) => DropdownMenuItem(value: b.id, child: Text(b.name))).toList(),
                    onChanged: (v) => setState(() => batchId = v),
                  ),
                const SizedBox(height: 12),
                if (coaches.data.isEmpty)
                  const Text('No coaches assigned to this activity.', style: TextStyle(color: Colors.orange))
                else
                  DropdownButtonFormField<String>(
                    initialValue: coachId,
                    decoration: const InputDecoration(labelText: 'Coach'),
                    items: coaches.data.map((c) => DropdownMenuItem(value: c.id, child: Text(c.name))).toList(),
                    onChanged: (v) => setState(() => coachId = v),
                  ),
                const SizedBox(height: 12),
                ListTile(
                  contentPadding: EdgeInsets.zero,
                  title: Text('Date: ${DateFormat.yMMMd().format(classDate)}'),
                  trailing: const Icon(Icons.calendar_today, size: 18),
                  onTap: () async {
                    final picked = await showDatePicker(
                      context: context,
                      initialDate: classDate,
                      firstDate: DateTime.now().subtract(const Duration(days: 365)),
                      lastDate: DateTime.now().add(const Duration(days: 365)),
                    );
                    if (picked != null) setState(() => classDate = picked);
                  },
                ),
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
                if (batchId == null || coachId == null) {
                  setState(() => errorText = 'A batch and coach are required.');
                  return;
                }
                final start = '${startTime.hour.toString().padLeft(2, '0')}:${startTime.minute.toString().padLeft(2, '0')}';
                final end = '${endTime.hour.toString().padLeft(2, '0')}:${endTime.minute.toString().padLeft(2, '0')}';
                try {
                  await ref.read(classRepositoryProvider).createClass(
                        batchId: batchId!,
                        activityId: activityId,
                        coachId: coachId!,
                        classDate: classDate,
                        startTime: start,
                        endTime: end,
                      );
                  ref.invalidate(recentClassesProvider);
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
