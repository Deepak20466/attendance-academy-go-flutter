import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/async_value_widget.dart';
import '../../auth/data/auth_controller.dart';
import '../../coaches/data/coach_repository.dart';
import '../data/class_repository.dart';
import '../data/substitution_repository.dart';

class SubstitutionsScreen extends ConsumerWidget {
  const SubstitutionsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final session = ref.watch(authControllerProvider).valueOrNull;
    final isAdmin = session?.isAdmin ?? false;
    final subs = ref.watch(isAdmin ? allSubstitutionsProvider : mySubstitutionsProvider);

    return Scaffold(
      appBar: AppBar(title: Text(isAdmin ? 'Substitutions' : 'My substitute classes')),
      floatingActionButton: isAdmin
          ? FloatingActionButton.extended(
              onPressed: () => _showCreateDialog(context, ref),
              icon: const Icon(Icons.add),
              label: const Text('Assign substitute'),
            )
          : null,
      body: RefreshIndicator(
        onRefresh: () => ref.refresh((isAdmin ? allSubstitutionsProvider : mySubstitutionsProvider).future),
        child: AsyncValueWidget(
          value: subs,
          data: (items) {
            if (items.isEmpty) {
              return ListView(
                children: [
                  Padding(
                    padding: const EdgeInsets.all(32),
                    child: Center(
                      child: Text(
                        isAdmin ? 'No substitutions recorded yet.' : 'You have no substitute assignments.',
                        style: const TextStyle(color: Colors.grey),
                      ),
                    ),
                  ),
                ],
              );
            }
            return ListView.separated(
              padding: const EdgeInsets.symmetric(vertical: 8),
              itemCount: items.length,
              separatorBuilder: (_, __) => const Divider(height: 1),
              itemBuilder: (context, index) {
                final sub = items[index];
                final active = sub.status == 'active';
                return _SubstitutionTile(
                  substitution: sub,
                  isAdmin: isAdmin,
                  onCancel: active
                      ? () async {
                          await ref.read(substitutionRepositoryProvider).cancel(sub.id);
                          ref.invalidate(allSubstitutionsProvider);
                        }
                      : null,
                  onMarkAttendance: !isAdmin && active
                      ? () => context.push('/classes/${sub.classId}/attendance')
                      : null,
                );
              },
            );
          },
        ),
      ),
    );
  }

  Future<void> _showCreateDialog(BuildContext context, WidgetRef ref) async {
    final classesResult = await ref.read(classRepositoryProvider).list(pageSize: 50);
    final coachesResult = await ref.read(coachRepositoryProvider).list(pageSize: 100);
    if (!context.mounted) return;

    if (classesResult.data.isEmpty || coachesResult.data.length < 2) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Need at least one class and two coaches to assign a substitute.')),
      );
      return;
    }

    String selectedClassId = classesResult.data.first.id;
    String selectedCoachId = coachesResult.data.first.id;
    final reasonController = TextEditingController();
    String? errorText;

    await showDialog<void>(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setState) => AlertDialog(
          title: const Text('Assign substitute'),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                DropdownButtonFormField<String>(
                  initialValue: selectedClassId,
                  decoration: const InputDecoration(labelText: 'Class'),
                  items: classesResult.data
                      .map((c) => DropdownMenuItem(
                            value: c.id,
                            child: Text('${DateFormat.yMMMd().format(c.classDate)} • ${c.startTime}-${c.endTime}'),
                          ))
                      .toList(),
                  onChanged: (v) => setState(() => selectedClassId = v!),
                ),
                const SizedBox(height: 12),
                DropdownButtonFormField<String>(
                  initialValue: selectedCoachId,
                  decoration: const InputDecoration(labelText: 'Substitute coach'),
                  items: coachesResult.data.map((c) => DropdownMenuItem(value: c.id, child: Text(c.name))).toList(),
                  onChanged: (v) => setState(() => selectedCoachId = v!),
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: reasonController,
                  decoration: const InputDecoration(labelText: 'Reason'),
                  maxLines: 2,
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
                try {
                  await ref.read(substitutionRepositoryProvider).create(
                        classId: selectedClassId,
                        substituteCoachId: selectedCoachId,
                        reason: reasonController.text.trim(),
                      );
                  ref.invalidate(allSubstitutionsProvider);
                  if (context.mounted) Navigator.pop(context);
                } catch (e) {
                  setState(() => errorText = 'Failed: $e');
                }
              },
              child: const Text('Assign'),
            ),
          ],
        ),
      ),
    );
  }
}

class _SubstitutionTile extends ConsumerWidget {
  const _SubstitutionTile({
    required this.substitution,
    required this.isAdmin,
    this.onCancel,
    this.onMarkAttendance,
  });

  final SubstitutionModel substitution;
  final bool isAdmin;
  final VoidCallback? onCancel;
  final VoidCallback? onMarkAttendance;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final active = substitution.status == 'active';
    final color = active ? StatusColors.approved : StatusColors.excused;

    return ListTile(
      title: Text(substitution.reason.isEmpty ? 'Substitution' : substitution.reason),
      subtitle: Text('Class ${substitution.classId.substring(0, 8)}…'),
      leading: Chip(
        label: Text(substitution.status, style: TextStyle(color: color, fontSize: 12)),
        backgroundColor: color.withValues(alpha: 0.12),
        visualDensity: VisualDensity.compact,
      ),
      trailing: Wrap(
        spacing: 4,
        children: [
          if (onMarkAttendance != null)
            IconButton(icon: const Icon(Icons.checklist_outlined), tooltip: 'Mark attendance', onPressed: onMarkAttendance),
          if (isAdmin && onCancel != null)
            IconButton(icon: const Icon(Icons.cancel_outlined), tooltip: 'Cancel', onPressed: onCancel),
        ],
      ),
    );
  }
}
