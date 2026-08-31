import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/widgets/async_value_widget.dart';
import '../../classes/data/class_repository.dart';
import '../data/student_models.dart';
import '../data/student_repository.dart';

class StudentsScreen extends ConsumerWidget {
  const StudentsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final students = ref.watch(studentsListProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Students')),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _showStudentFormDialog(context, ref),
        icon: const Icon(Icons.add),
        label: const Text('Add student'),
      ),
      body: RefreshIndicator(
        onRefresh: () => ref.refresh(studentsListProvider.future),
        child: AsyncValueWidget(
          value: students,
          onRetry: () => ref.invalidate(studentsListProvider),
          data: (result) {
            if (result.data.isEmpty) {
              return ListView(
                children: const [
                  Padding(
                    padding: EdgeInsets.all(32),
                    child: Center(
                      child: Text(
                        'No students found.',
                        style: TextStyle(color: Colors.grey),
                      ),
                    ),
                  ),
                ],
              );
            }
            return ListView.separated(
              padding: const EdgeInsets.symmetric(vertical: 8),
              itemCount: result.data.length,
              separatorBuilder: (_, __) => const Divider(height: 1),
              itemBuilder: (context, index) {
                final student = result.data[index];
                return ListTile(
                  leading: CircleAvatar(
                    child: Text(
                      student.name.isNotEmpty
                          ? student.name[0].toUpperCase()
                          : '?',
                    ),
                  ),
                  title: Text(
                    student.name,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                  subtitle: Text(
                    student.guardianPhone.isNotEmpty
                        ? student.guardianPhone
                        : student.phone,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                  onTap: () => context.push('/students/${student.id}'),
                  trailing: PopupMenuButton<String>(
                    tooltip: 'Student actions',
                    onSelected: (action) {
                      if (action == 'edit') {
                        _showStudentFormDialog(context, ref, existing: student);
                      } else if (action == 'remove') {
                        _confirmAndRemove(context, ref, student);
                      }
                    },
                    itemBuilder: (context) => const [
                      PopupMenuItem(value: 'edit', child: Text('Edit')),
                      PopupMenuItem(value: 'remove', child: Text('Remove')),
                    ],
                  ),
                );
              },
            );
          },
        ),
      ),
    );
  }

  Future<void> _confirmAndRemove(
    BuildContext context,
    WidgetRef ref,
    StudentModel student,
  ) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Remove student?'),
        content: Text(
          "Remove ${student.name} from the active roster? Their record and attendance history are kept, "
          "but they'll no longer appear in student lists.",
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            style: FilledButton.styleFrom(backgroundColor: Colors.red),
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Remove'),
          ),
        ],
      ),
    );
    if (confirmed != true || !context.mounted) return;

    final messenger = ScaffoldMessenger.of(context);
    messenger.clearSnackBars();
    try {
      await ref
          .read(studentRepositoryProvider)
          .update(
            id: student.id,
            batchId: student.batchId,
            name: student.name,
            phone: student.phone,
            guardianName: student.guardianName,
            guardianPhone: student.guardianPhone,
            email: student.email,
            dateOfBirth: student.dateOfBirth,
            isActive: false,
          );
      ref.invalidate(studentsListProvider);
      ref.invalidate(studentDetailProvider(student.id));
      messenger.showSnackBar(
        SnackBar(
          content: Text('${student.name} removed from the active roster.'),
        ),
      );
    } catch (e) {
      messenger.showSnackBar(
        SnackBar(
          content: Text('Failed to remove student: ${_friendlyError(e)}'),
        ),
      );
    }
  }

  Future<void> _showStudentFormDialog(
    BuildContext context,
    WidgetRef ref, {
    StudentModel? existing,
  }) async {
    final isEdit = existing != null;
    final batches = await ref.read(classRepositoryProvider).listBatches();
    if (!context.mounted) return;

    if (batches.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Create a batch first before adding students.'),
        ),
      );
      return;
    }

    final nameController = TextEditingController(text: existing?.name ?? '');
    final phoneController = TextEditingController(text: existing?.phone ?? '');
    final guardianNameController = TextEditingController(
      text: existing?.guardianName ?? '',
    );
    final guardianPhoneController = TextEditingController(
      text: existing?.guardianPhone ?? '',
    );
    String? selectedBatchId = existing?.batchId ?? batches.first.id;
    // The batch a student currently belongs to might not be in this
    // activity-scoped dropdown (e.g. it was reassigned since); fall back to
    // the first batch rather than leaving the dropdown with a value it has
    // no matching item for, which would throw.
    if (!batches.any((b) => b.id == selectedBatchId)) {
      selectedBatchId = batches.first.id;
    }
    String? errorText;
    bool isSaving = false;

    await showDialog<void>(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setState) => AlertDialog(
          title: Text(isEdit ? 'Edit student' : 'Add student'),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                TextField(
                  controller: nameController,
                  decoration: const InputDecoration(labelText: 'Name'),
                ),
                const SizedBox(height: 12),
                DropdownButtonFormField<String>(
                  initialValue: selectedBatchId,
                  decoration: const InputDecoration(labelText: 'Batch'),
                  items: batches
                      .map(
                        (b) =>
                            DropdownMenuItem(value: b.id, child: Text(b.name)),
                      )
                      .toList(),
                  onChanged: (v) => setState(() => selectedBatchId = v),
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: phoneController,
                  decoration: const InputDecoration(labelText: 'Phone'),
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: guardianNameController,
                  decoration: const InputDecoration(labelText: 'Guardian name'),
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: guardianPhoneController,
                  decoration: const InputDecoration(
                    labelText: 'Guardian phone',
                  ),
                ),
                if (errorText != null) ...[
                  const SizedBox(height: 12),
                  Text(errorText!, style: const TextStyle(color: Colors.red)),
                ],
              ],
            ),
          ),
          actions: [
            TextButton(
              onPressed: isSaving ? null : () => Navigator.pop(context),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: isSaving
                  ? null
                  : () async {
                      if (nameController.text.trim().isEmpty ||
                          selectedBatchId == null) {
                        setState(
                          () => errorText = 'Name and batch are required.',
                        );
                        return;
                      }
                      setState(() {
                        isSaving = true;
                        errorText = null;
                      });
                      try {
                        if (isEdit) {
                          await ref
                              .read(studentRepositoryProvider)
                              .update(
                                id: existing.id,
                                batchId: selectedBatchId!,
                                name: nameController.text.trim(),
                                phone: phoneController.text.trim(),
                                guardianName: guardianNameController.text
                                    .trim(),
                                guardianPhone: guardianPhoneController.text
                                    .trim(),
                                email: existing.email,
                                dateOfBirth: existing.dateOfBirth,
                                isActive: existing.isActive,
                              );
                          ref.invalidate(studentDetailProvider(existing.id));
                        } else {
                          await ref
                              .read(studentRepositoryProvider)
                              .create(
                                batchId: selectedBatchId!,
                                name: nameController.text.trim(),
                                phone: phoneController.text.trim(),
                                guardianName: guardianNameController.text
                                    .trim(),
                                guardianPhone: guardianPhoneController.text
                                    .trim(),
                              );
                        }
                        ref.invalidate(studentsListProvider);
                        if (context.mounted) Navigator.pop(context);
                      } catch (e) {
                        setState(() {
                          isSaving = false;
                          errorText =
                              'Failed to ${isEdit ? 'update' : 'create'} student: ${_friendlyError(e)}';
                        });
                      }
                    },
              child: isSaving
                  ? const SizedBox(
                      height: 18,
                      width: 18,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : Text(isEdit ? 'Save' : 'Add'),
            ),
          ],
        ),
      ),
    );
  }
}

// Mirrors AsyncValueWidget's translation so a raw DioException (CORS/client
// internals, meaningless to the person tapping the button) never reaches a
// snackbar; only the server's own error message or a plain fallback does.
String _friendlyError(Object error) {
  if (error is DioException) {
    switch (error.type) {
      case DioExceptionType.connectionError:
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.receiveTimeout:
      case DioExceptionType.sendTimeout:
        return "You're offline. Check your connection and try again.";
      default:
        final data = error.response?.data;
        final serverMessage = data is Map ? data['error'] as String? : null;
        return serverMessage ?? 'Something went wrong. Please try again.';
    }
  }
  return 'Something went wrong. Please try again.';
}
