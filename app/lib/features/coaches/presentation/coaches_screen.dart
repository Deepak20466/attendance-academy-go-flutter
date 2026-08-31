import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/widgets/async_value_widget.dart';
import '../../activities/data/activity_repository.dart';
import '../data/coach_repository.dart';

class CoachesScreen extends ConsumerWidget {
  const CoachesScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final coaches = ref.watch(_coachesListProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Coaches')),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _showCreateDialog(context, ref),
        icon: const Icon(Icons.add),
        label: const Text('Add coach'),
      ),
      body: RefreshIndicator(
        onRefresh: () => ref.refresh(_coachesListProvider.future),
        child: AsyncValueWidget(
          value: coaches,
          data: (result) {
            if (result.data.isEmpty) {
              return ListView(
                children: const [
                  Padding(
                    padding: EdgeInsets.all(32),
                    child: Center(child: Text('No coaches found.', style: TextStyle(color: Colors.grey))),
                  ),
                ],
              );
            }
            return ListView.separated(
              padding: const EdgeInsets.symmetric(vertical: 8),
              itemCount: result.data.length,
              separatorBuilder: (_, __) => const Divider(height: 1),
              itemBuilder: (context, index) {
                final coach = result.data[index];
                return ListTile(
                  leading: CircleAvatar(child: Text(coach.name.isNotEmpty ? coach.name[0].toUpperCase() : '?')),
                  title: Text(coach.name),
                  subtitle: Text('${coach.employeeCode} • ${coach.activityIds.length} activit${coach.activityIds.length == 1 ? 'y' : 'ies'}'),
                  trailing: coach.isActive ? null : const Chip(label: Text('Inactive')),
                );
              },
            );
          },
        ),
      ),
    );
  }

  Future<void> _showCreateDialog(BuildContext context, WidgetRef ref) async {
    final activities = await ref.read(activityRepositoryProvider).list();
    if (!context.mounted) return;
    if (activities.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Create an activity first.')),
      );
      return;
    }

    final nameController = TextEditingController();
    final emailController = TextEditingController();
    final phoneController = TextEditingController();
    final passwordController = TextEditingController();
    final employeeCodeController = TextEditingController();
    final salaryController = TextEditingController();
    final selectedActivityIds = <String>{};
    String? errorText;

    await showDialog<void>(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setState) => AlertDialog(
          title: const Text('Add coach'),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                TextField(controller: nameController, decoration: const InputDecoration(labelText: 'Name')),
                const SizedBox(height: 12),
                TextField(controller: emailController, decoration: const InputDecoration(labelText: 'Email')),
                const SizedBox(height: 12),
                TextField(controller: phoneController, decoration: const InputDecoration(labelText: 'Phone')),
                const SizedBox(height: 12),
                TextField(
                  controller: passwordController,
                  obscureText: true,
                  decoration: const InputDecoration(labelText: 'Password (min 8 characters)'),
                ),
                const SizedBox(height: 12),
                TextField(controller: employeeCodeController, decoration: const InputDecoration(labelText: 'Employee code')),
                const SizedBox(height: 12),
                TextField(
                  controller: salaryController,
                  keyboardType: TextInputType.number,
                  decoration: const InputDecoration(labelText: 'Monthly salary'),
                ),
                const SizedBox(height: 12),
                const Align(alignment: Alignment.centerLeft, child: Text('Activities')),
                const SizedBox(height: 4),
                Wrap(
                  spacing: 6,
                  children: activities.map((a) {
                    return FilterChip(
                      label: Text(a.name),
                      selected: selectedActivityIds.contains(a.id),
                      onSelected: (sel) => setState(() => sel ? selectedActivityIds.add(a.id) : selectedActivityIds.remove(a.id)),
                    );
                  }).toList(),
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
                final salary = double.tryParse(salaryController.text);
                if (nameController.text.trim().isEmpty ||
                    emailController.text.trim().isEmpty ||
                    employeeCodeController.text.trim().isEmpty ||
                    passwordController.text.length < 8 ||
                    salary == null) {
                  setState(() => errorText = 'Name, email, employee code, an 8+ character password and a valid salary are required.');
                  return;
                }
                try {
                  await ref.read(coachRepositoryProvider).create(
                        name: nameController.text.trim(),
                        email: emailController.text.trim(),
                        phone: phoneController.text.trim(),
                        password: passwordController.text,
                        employeeCode: employeeCodeController.text.trim(),
                        monthlySalary: salary,
                        activityIds: selectedActivityIds.toList(),
                      );
                  ref.invalidate(_coachesListProvider);
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

final _coachesListProvider = FutureProvider.autoDispose((ref) {
  return ref.watch(coachRepositoryProvider).list();
});
