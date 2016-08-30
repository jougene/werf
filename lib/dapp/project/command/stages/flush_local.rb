module Dapp
  # Project
  class Project
    # Command
    module Command
      module Stages
        # FlushLocal
        module FlushLocal
          def stages_flush_local
            build_configs.map(&:_basename).uniq.each do |basename|
              lock("#{basename}.images") do
                log_step_with_indent(basename) do
                  project_containers_flush(basename)
                  remove_images_by_query(project_images(basename))
                end
              end
            end
          end
        end
      end
    end
  end # Project
end # Dapp
